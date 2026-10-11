package ytnode

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/pion/rtp"
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
	currentReceiver := ""
	currentConfigIdx := -1

	emit := func(buf []byte) {
		yt.rtcPacketsInc <- bytes.Clone(buf)
	}

	var rxPkts, rxBytes, rxLost int
	var lastSeq uint16
	lastStats := time.Now()

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

		if gap := packet.SequenceNumber - lastSeq - 1; rxPkts > 0 && gap < 0x8000 {
			rxLost += int(gap)
		}
		lastSeq = packet.SequenceNumber
		rxPkts++
		rxBytes += len(packet.Payload)
		if time.Since(lastStats) > 5*time.Second {
			yt.log.Debugf(
				"RTP rx %v: %.2f Mbit/s, %d pkts, %d lost",
				mid, float64(rxBytes)*8/1e6/time.Since(lastStats).Seconds(),
				rxPkts, rxLost,
			)
			lastStats = time.Now()
			rxPkts, rxBytes, rxLost = 0, 0, 0
		}

		if !extractor.Extract(packet, emit) {
			yt.log.Warnf("Received malformed packet\n")
		}
	}
}

const (
	keepAliveDelay = 3 * time.Second

	// The SFU drops frames if there are too many of them (somewhere above
	// 250 fps), so packets that come in faster than that are batched.
	frameInterval = 5 * time.Millisecond

	// The SFU holds the last received RTP packet until the next one arrives, so
	// when the queue runs empty, an empty frame is sent to push the data out.
	chaserDelay = frameInterval
)

// StreamToDataTrack sends IP packets as soon as possible: everything that is
// queued at the moment goes into one frame.
func (yt *YTClient) StreamToDataTrack(track *webrtc.TrackLocalStaticRTP) {
	yt.log.Infof("Started streaming to '%v'\n", yt.targetName)

	constructor := NewRTPConstructor()
	write := func(p *rtp.Packet) error {
		return track.WriteRTP(p)
	}

	keepAliveTimer := time.NewTimer(keepAliveDelay)
	chaserTimer := time.NewTimer(chaserDelay)
	chaserPending := true
	lastFrame := time.Now()

	// sendFrame sends buf and everything else that is queued as one frame
	sendFrame := func(buf []byte) error {
		constructor.BeginFrame()
		if buf != nil {
			constructor.Add(buf)
		}

	drain:
		for constructor.FramePackets() < yt.batching.MaxBatchSize {
			select {
			case buf = <-yt.rtcPacketsOut:
				constructor.Add(buf)
			default:
				break drain
			}
		}

		lastFrame = time.Now()
		return constructor.EndFrame(write)
	}

	for {
		var err error

		select {
		case <-keepAliveTimer.C:
			err = track.WriteRTP(constructor.NewDummyPacket())
			keepAliveTimer.Reset(keepAliveDelay)

		case <-chaserTimer.C:
			if chaserPending {
				chaserPending = false
				err = sendFrame(nil)
			}

		case buf := <-yt.rtcPacketsOut:
			// let more packets queue up if the previous frame was just sent
			if wait := frameInterval - time.Since(lastFrame); wait > 0 {
				time.Sleep(wait)
			}

			err = sendFrame(buf)
			chaserPending = true
			chaserTimer.Reset(chaserDelay)
			keepAliveTimer.Reset(keepAliveDelay)
		}

		if err != nil {
			yt.log.Errorf("Error writing to track: %v\n", err)
			return
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
