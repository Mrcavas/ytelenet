package ytnode

import (
	"fmt"
	"regexp"

	"github.com/google/uuid"
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

	if _, err := pc.AddTrack(audioTrack); err != nil {
		return fmt.Errorf("failed to add publisher audio track: %w", err)
	}

	if _, err := pc.AddTrack(dataTrack); err != nil {
		return fmt.Errorf("failed to add publisher data track: %w", err)
	}

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
