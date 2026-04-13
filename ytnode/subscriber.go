package ytnode

import (
	"fmt"
	"regexp"

	"github.com/pion/webrtc/v4"
)

func (yt *YTClient) InitRTCSubscriber(
	onVideoTrack func(mid string, track *webrtc.TrackRemote),
) error {
	yt.log.Infof("Initializing RTC subscriber\n")

	iceServersData := yt.ServerHello.RtcConfiguration.IceServers
	iceServers := make([]webrtc.ICEServer, len(iceServersData))

	for i, ice := range iceServersData {
		iceServers[i] = webrtc.ICEServer{
			URLs:       ice.Urls,
			Username:   ice.Username,
			Credential: ice.Credential,
		}
	}

	m, err := makeMediaEngine()
	if err != nil {
		return err
	}
	api := webrtc.NewAPI(webrtc.WithMediaEngine(m))

	pc, err := api.NewPeerConnection(
		webrtc.Configuration{
			ICEServers: iceServers,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create subscriber connection: %w", err)
	}

	yt.rtcSubscriber = pc

	pc.OnTrack(
		func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
			mid := receiver.RTPTransceiver().Mid()
			codecName := track.Codec().MimeType
			yt.log.Debugf("Got incoming track! Codec: %s, MID: %s", codecName, mid)

			if codecName == webrtc.MimeTypeVP8 {
				onVideoTrack(mid, track)
				return
			}

			go DrainTrack(track)
		},
	)

	pc.OnICECandidate(
		func(candidate *webrtc.ICECandidate) {
			if candidate == nil {
				yt.log.Debugf("Subscriber ICE gathering complete\n")
				return
			}

			init := candidate.ToJSON()

			msg := WSMessageOutgoing{
				WebrtcIceCandidate: &WebrtcIceCandidatePayload{
					Candidate:        init.Candidate,
					SdpMid:           *init.SDPMid,
					UsernameFragment: *yt.subscriberUfrag,
					SdpMlineIndex:    int(*init.SDPMLineIndex),
					Target:           "SUBSCRIBER",
					PcSeq:            1,
				},
			}

			yt.SendWS(&msg)
		},
	)

	pc.OnICEConnectionStateChange(
		func(state webrtc.ICEConnectionState) {
			yt.log.Infof("Subscriber ICE Connection State: %s", state)
		},
	)

	return nil
}

func makeMediaEngine() (*webrtc.MediaEngine, error) {
	m := &webrtc.MediaEngine{}

	if err := m.RegisterCodec(
		webrtc.RTPCodecParameters{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:  webrtc.MimeTypeOpus,
				ClockRate: 48000,
				Channels:  2,
			},
			PayloadType: 111,
		}, webrtc.RTPCodecTypeAudio,
	); err != nil {
		return nil, fmt.Errorf("failed to register opus: %w", err)
	}

	if err := m.RegisterCodec(
		webrtc.RTPCodecParameters{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:  webrtc.MimeTypeVP8,
				ClockRate: 90000,
			},
			PayloadType: 96,
		}, webrtc.RTPCodecTypeVideo,
	); err != nil {
		return nil, fmt.Errorf("failed to register vp8: %w", err)
	}

	return m, nil
}

func (yt *YTClient) HandleRTCSubscriberOffer(offer string) error {
	if yt.rtcSubscriber == nil {
		return fmt.Errorf("subscriber PC is nil")
	}

	err := yt.rtcSubscriber.SetRemoteDescription(
		webrtc.SessionDescription{
			Type: webrtc.SDPTypeOffer,
			SDP:  offer,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to set subscriber remote description: %w", err)
	}

	answer, err := yt.rtcSubscriber.CreateAnswer(nil)
	if err != nil {
		return fmt.Errorf("failed to create subscriber answer: %w", err)
	}

	ufragRegex := regexp.MustCompile(`a=ice-ufrag:([a-zA-Z0-9+/]+)`)
	matches := ufragRegex.FindStringSubmatch(answer.SDP)
	if len(matches) > 1 {
		yt.subscriberUfrag = &matches[1]
		yt.log.Debugf("Extracted subscriber ICE ufrag: %s", *yt.subscriberUfrag)
	} else {
		return fmt.Errorf("couldn't find ice-ufrag in generated SDP")
	}

	if err := yt.rtcSubscriber.SetLocalDescription(answer); err != nil {
		return fmt.Errorf("failed to set subscriber local description: %w", err)
	}

	yt.SendWS(
		&WSMessageOutgoing{
			SubscriberSdpAnswer: &SdpPayload{
				PcSeq: 1,
				Sdp:   answer.SDP,
			},
		},
	)

	return nil
}

func (yt *YTClient) HandleIncomingSubscriberCandidate(cand *WebrtcIceCandidatePayload) error {
	if yt.rtcSubscriber == nil {
		return fmt.Errorf("subscriber PC is nil")
	}

	sdpMLineIndex := uint16(cand.SdpMlineIndex)

	return yt.rtcSubscriber.AddICECandidate(
		webrtc.ICECandidateInit{
			Candidate:     cand.Candidate,
			SDPMid:        &cand.SdpMid,
			SDPMLineIndex: &sdpMLineIndex,
		},
	)
}

func (yt *YTClient) CloseRTCSubscriber() error {
	if yt.rtcSubscriber == nil {
		return nil
	}

	return yt.rtcSubscriber.GracefulClose()
}
