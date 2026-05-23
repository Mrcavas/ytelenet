package ytnode

import (
  "fmt"
  "regexp"

  "github.com/google/uuid"
  "github.com/pion/rtcp"
  "github.com/pion/webrtc/v4"
)

func (yt *YTClient) InitRTCPublisher(
  dataTrack webrtc.TrackLocal, onStateChange func(webrtc.ICEConnectionState),
) error {
  yt.log.Infof("Initializing RTC publisher\n")

  iceServersData := yt.ServerHello.RtcConfiguration.IceServers
  iceServers := make([]webrtc.ICEServer, len(iceServersData))

  for i, ice := range iceServersData {
    iceServers[i] = webrtc.ICEServer{
      URLs:       ice.Urls,
      Username:   ice.Username,
      Credential: ice.Credential,
    }
  }

  pc, err := webrtc.NewPeerConnection(
    webrtc.Configuration{
      ICEServers: iceServers,
    },
  )
  if err != nil {
    return fmt.Errorf("failed to create publisher connection: %w", err)
  }

  yt.rtcPublisher = pc

  audioTrack, err := webrtc.NewTrackLocalStaticSample(
    webrtc.RTPCodecCapability{
      MimeType: webrtc.MimeTypeOpus,
    },
    "{"+uuid.NewString()+"}",
    "-",
  )
  if err != nil {
    return fmt.Errorf("failed to create publisher audio track: %w", err)
  }

  audioRtpSender, err := pc.AddTrack(audioTrack)
  if err != nil {
    return fmt.Errorf("failed to add publisher audio track: %w", err)
  }
  go yt.DrainRTP(audioRtpSender)

  dataRtpSender, err := pc.AddTrack(dataTrack)
  if err != nil {
    return fmt.Errorf("failed to add publisher data track: %w", err)
  }
  go yt.DrainRTP(dataRtpSender)

  pc.OnICECandidate(
    func(candidate *webrtc.ICECandidate) {
      if candidate == nil {
        yt.log.Debugf("Publisher ICE gathering complete\n")
        return
      }

      init := candidate.ToJSON()

      msg := WSMessageOutgoing{
        WebrtcIceCandidate: &WebrtcIceCandidatePayload{
          Candidate:        init.Candidate,
          SdpMid:           *init.SDPMid,
          UsernameFragment: *yt.publisherUfrag,
          SdpMlineIndex:    int(*init.SDPMLineIndex),
          Target:           "PUBLISHER",
          PcSeq:            1,
        },
      }

      yt.SendWS(&msg)
    },
  )

  pc.OnICEConnectionStateChange(onStateChange)

  offer, err := pc.CreateOffer(nil)
  if err != nil {
    return fmt.Errorf("failed to create publisher offer: %w", err)
  }

  ufragRegex := regexp.MustCompile(`a=ice-ufrag:([a-zA-Z0-9+/]+)`)
  matches := ufragRegex.FindStringSubmatch(offer.SDP)
  if len(matches) > 1 {
    yt.publisherUfrag = &matches[1]
    yt.log.Debugf("Extracted ICE ufrag: %s", *yt.publisherUfrag)
  } else {
    return fmt.Errorf("couldn't find ice-ufrag in generated SDP")
  }

  if err := pc.SetLocalDescription(offer); err != nil {
    return fmt.Errorf("failed to set publisher local description: %w", err)
  }

  offerPayload, err := GetDefaultPayload[PublisherSdpOfferPayload]("publisherSdpOffer")
  if err != nil {
    return fmt.Errorf("failed to get publisherSdpOffer payload: %w", err)
  }

  offerPayload.Sdp = offer.SDP
  yt.SendWS(
    &WSMessageOutgoing{
      PublisherSdpOffer: offerPayload,
    },
  )

  return nil
}

func (yt *YTClient) HandleRTCPublisherAnswer(sdp string) error {
  if yt.rtcPublisher == nil {
    return fmt.Errorf("publisher PC is nil")
  }

  return yt.rtcPublisher.SetRemoteDescription(
    webrtc.SessionDescription{
      Type: webrtc.SDPTypeAnswer,
      SDP:  sdp,
    },
  )
}

func (yt *YTClient) HandleIncomingPublisherCandidate(cand *WebrtcIceCandidatePayload) error {
  if yt.rtcPublisher == nil {
    return fmt.Errorf("publisher PC is nil")
  }

  sdpMLineIndex := uint16(cand.SdpMlineIndex)

  return yt.rtcPublisher.AddICECandidate(
    webrtc.ICECandidateInit{
      Candidate:     cand.Candidate,
      SDPMid:        &cand.SdpMid,
      SDPMLineIndex: &sdpMLineIndex,
    },
  )
}

func (yt *YTClient) CloseRTCPublisher() error {
  if yt.rtcPublisher == nil {
    return nil
  }

  return yt.rtcPublisher.GracefulClose()
}

func (yt *YTClient) DrainRTP(rtpSender *webrtc.RTPSender) {
  rtcpBuf := make([]byte, 1500)
  for {
    // Continuously read from the sender to prevent buffer buildup
    n, _, rtcpErr := rtpSender.Read(rtcpBuf)
    if rtcpErr != nil {
      return
    }

    if n > 0 {
      // Parse the raw bytes into actual RTCP packets
      packets, err := rtcp.Unmarshal(rtcpBuf[:n])
      if err != nil {
        yt.log.Errorf("Error unmarshalling rtcp: %v", err)
        continue
      }

      for _, pkt := range packets {
        switch p := pkt.(type) {
        case *rtcp.ReceiverEstimatedMaximumBitrate:
          // THIS IS THE SMOKING GUN.
          // The SFU is explicitly telling us our bandwidth limit!
          mbps := float64(p.Bitrate) / 1000000.0
          yt.log.Infof("SFU REMB Command: Limit bandwidth to %.2f Mbps", mbps)

        case *rtcp.TransportLayerCC:
          // TWCC packets tell us if the SFU thinks we are causing network congestion
          yt.log.Debugf("SFU TWCC: Congestion report received")

        case *rtcp.PictureLossIndication:
          // PLI means the SFU failed to decode a frame and wants a new Keyframe
          yt.log.Debugf("SFU PLI: Frame dropped, requesting keyframe")
        }
      }
    }
  }
}
