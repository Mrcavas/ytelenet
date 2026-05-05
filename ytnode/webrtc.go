package ytnode

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v4"
)

func (yt *YTClient) InitializeRTC() error {
	yt.log.Infof("Initializing WebRTC\n")
	slots, err := GetDefaultPayload[SetSlotsPayload]("setSlots")
	if err != nil {
		return err

	}

	slots.Key = 1
	yt.SendWS(
		&WSMessageOutgoing{
			SetSlots: slots,
		},
	)

	slots.Key = 2
	yt.SendWS(
		&WSMessageOutgoing{
			SetSlots: slots,
		},
	)

	track, err := MakeDataTrack()
	if err != nil {
		return err

	}

	onStateChange := func(state webrtc.ICEConnectionState) {
		yt.log.Warnf("Publisher ICE Connection State: %s", state)

		if state == webrtc.ICEConnectionStateConnected {
			go yt.StreamToDataTrack(track)
			yt.publisherConnected <- struct{}{}
		}
	}

	if err := yt.InitRTCPublisher(track, onStateChange); err != nil {
		return fmt.Errorf("couldn't initialize WebRTC publisher: %w", err)
	}

	onVideoTrack := func(mid string, track *webrtc.TrackRemote) {
		go yt.StreamFromDataTrack(mid, track)
	}

	if err := yt.InitRTCSubscriber(onVideoTrack); err != nil {
		return fmt.Errorf("couldn't initialize WebRTC subscriber: %w", err)
	}

	return nil
}

func MakeDataTrack() (*webrtc.TrackLocalStaticRTP, error) {
	videoTrack, err := webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{
			MimeType: webrtc.MimeTypeVP8,
		},
		"{"+uuid.NewString()+"}",
		"-",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create publisher video track: %w", err)
	}

	return videoTrack, nil
}

func (yt *YTClient) StreamFromDataTrack(mid string, track *webrtc.TrackRemote) {
	yt.log.Infof("Started streaming from %v (%v)\n", mid, yt.targetName)

	extractor := NewRTPExtractor()
	yt.peerConfigMu.RLock()
	currentReceiver := ""
	currentConfigIdx := -1
	yt.peerConfigMu.RUnlock()

	for {
		packet, _, err := track.ReadRTP()
		if errors.Is(err, io.EOF) {
			yt.log.Infof("Data track finished\n")
			return
		}

		if err != nil {
			yt.log.Errorf("Error reading track: %v\n", err)
			return
		}

		yt.peerConfigMu.RLock()
		if currentConfigIdx != yt.peerConfigUpdateIdx {
			currentConfigIdx = yt.peerConfigUpdateIdx

			id, ok := yt.peerMids.GetInverse(mid)
			if ok {
				currentReceiver, ok = yt.peerNames.Get(id)
			}

			if !ok {
				currentReceiver = ""
			}
			yt.log.Infof("Track info updated, now mid %v is %v", mid, currentReceiver)
		}
		yt.peerConfigMu.RUnlock()

		if currentReceiver != yt.targetName {
			continue
		}

		buf, ok := extractor.Extract(packet)
		if ok {
			if len(buf) > 0 {
				yt.rtcPacketsInc <- buf
			}
		} else {
			yt.log.Warnf("Received malformed packet\n")
		}
	}
}

func (yt *YTClient) StreamToDataTrack(track *webrtc.TrackLocalStaticRTP) {
	yt.log.Infof("Started streaming to '%v'\n", yt.targetName)

	constructor := NewRTPConstructor()
	keepAliveTimer := time.NewTimer(3 * time.Second)

	for {
		select {
		case <-keepAliveTimer.C:
			if err := track.WriteRTP(constructor.NewDummyPacket()); err != nil {
				yt.log.Errorf("Error writing dummy to track: %v\n", err)
				return
			}
			keepAliveTimer.Reset(time.Second * 3)

		case buf := <-yt.rtcPacketsOut:
			for packet := range constructor.NewPackets(buf) {
				if err := track.WriteRTP(packet); err != nil {
					yt.log.Errorf("Error writing to track: %v\n", err)
					return
				}
			}
			keepAliveTimer.Reset(time.Millisecond * 2)
		}
	}
}

func DrainTrack(track *webrtc.TrackRemote) {
	for {
		_, _, err := track.ReadRTP()
		if err != nil {
			return
		}
	}
}
