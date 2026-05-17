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
	slots1, err := GetDefaultPayload[SetSlotsPayload]("setSlots1")
	if err != nil {
		return err

	}
	slots2, err := GetDefaultPayload[SetSlotsPayload]("setSlots2")
	if err != nil {
		return err

	}

	yt.SendWS(
		&WSMessageOutgoing{
			SetSlots: slots1,
		},
	)
	yt.SendWS(
		&WSMessageOutgoing{
			SetSlots: slots2,
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

// func (yt *YTClient) StreamToDataTrack(track *webrtc.TrackLocalStaticRTP) {
//   yt.log.Infof("Started streaming REAL video to '%v'\n", yt.targetName)
//
//   videoFilePath := "video.ivf"
//
//   // Set up Pion's native RTP packetizer for VP8
//   payloader := &codecs.VP8Payloader{
//     EnablePictureID: true,
//   }
//   sequencer := rtp.NewRandomSequencer()
//
//   // RTP timestamp clock rate for VP8 is always 90,000 Hz
//   packetizer := rtp.NewPacketizer(
//     1200, // MTU
//     96,   // Payload Type
//     0,    // SSRC
//     payloader,
//     sequencer,
//     90000,
//   )
//
//   // Loop the video infinitely
//   for {
//     err := streamIVF(videoFilePath, track, packetizer, yt)
//     if err != nil {
//       yt.log.Errorf("Video stream loop ended/error: %v. Restarting...", err)
//       time.Sleep(1 * time.Second)
//     }
//   }
// }
//
// // streamIVF plays the file once until EOF, then returns nil
// func streamIVF(
//   filePath string, track *webrtc.TrackLocalStaticRTP, packetizer rtp.Packetizer,
//   yt *YTClient,
// ) error {
//   file, err := os.Open(filePath)
//   if err != nil {
//     return err
//   }
//   defer file.Close()
//
//   ivf, _, err := ivfreader.NewWith(file)
//   if err != nil {
//     return err
//   }
//
//   // 30 FPS = ~33.3 milliseconds per frame
//   ticker := time.NewTicker(time.Millisecond * 33)
//   defer ticker.Stop()
//
//   for {
//     frameBytes, _, err := ivf.ParseNextFrame()
//     if errors.Is(err, io.EOF) {
//       return nil // End of file, return so outer loop can restart
//     }
//     if err != nil {
//       return err
//     }
//
//     <-ticker.C // Wait for the next 30fps tick
//
//     // Packetize the frame.
//     // The second parameter is the amount of 90kHz ticks to advance the timestamp.
//     // For 30 FPS: 90,000 / 30 = 3000 ticks per frame.
//     packets := packetizer.Packetize(frameBytes, 3000)
//
//     // Write all resulting RTP packets for this frame to the track
//     for _, p := range packets {
//       if err := track.WriteRTP(p); err != nil {
//         return err
//       }
//     }
//   }
// }

func DrainTrack(track *webrtc.TrackRemote) {
	for {
		_, _, err := track.ReadRTP()
		if err != nil {
			return
		}
	}
}
