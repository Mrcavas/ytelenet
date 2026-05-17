package ytnode

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type Node struct {
	Log *logrus.Logger
	yt  *YTClient

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	events chan State
}

func MakeNew(log *logrus.Logger, roomUrl, name, target string) (
	*Node, error,
) {
	node := &Node{
		events: make(chan State, 2),
	}

	if err := InitDefaultPayloads(); err != nil {
		return nil, fmt.Errorf("failed to initialize default payloads: %w", err)
	}

	if err := InitDummyFrames(); err != nil {
		return nil, fmt.Errorf("failed to initialize dummy frames: %w", err)
	}

	yt := NewYTClient(log, roomUrl, name, target)
	node.yt = yt

	if err := yt.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize node '%v': %w", roomUrl, err)
	}

	_, err := yt.RequestStatesForNewPeer(yt.InitData.PeerID)
	if err != nil {
		return nil, fmt.Errorf("failed to request initial states: %w", err)
	}

	httpPinger := time.NewTicker(time.Duration(yt.InitData.ClientConfiguration.StateCheckIntervalSeconds) * time.Second)

	if err := yt.InitializeWS(); err != nil {
		return nil, fmt.Errorf("failed to initialize WS: %w", err)
	}
	yt.SendHello()

	node.ctx, node.cancel = context.WithCancel(context.Background())

	node.wg.Go(
		func() {
			var wsPinger *time.Ticker
			var wsPingerChan <-chan time.Time

			defer func() {
				log.Infof("Closing things")
				node.events <- StoppedState

				httpPinger.Stop()
				_ = yt.CloseWS()

				if wsPinger != nil {
					wsPinger.Stop()
					wsPinger = nil
					wsPingerChan = nil
				}

				if err := yt.CloseRTCPublisher(); err != nil {
					log.Errorf("Couldn't close WebRTC publisher: %v\n", err)
				}

				if err := yt.CloseRTCSubscriber(); err != nil {
					log.Errorf("Couldn't close WebRTC subscriber: %v\n", err)
				}
			}()

			for {
				select {
				case <-httpPinger.C:
					_, err := yt.PingStates()
					if err != nil {
						log.Errorf("Error pinging request-states: %v\n", err)
					}

				case <-wsPingerChan:
					yt.SendPing()

				case <-yt.wsInitialized:
					wsPinger = time.NewTicker(time.Duration(yt.ServerHello.PingPongConfiguration.PingInterval) * time.Millisecond)
					wsPingerChan = wsPinger.C

					err := yt.InitializeRTC()
					if err != nil {
						log.Errorf("Error initializing WebRTC: %v\n", err)
						return
					}

				case msg := <-yt.wsMessages:
					if err := yt.handleWSMessage(msg); err != nil {
						log.Errorf("Error handling WS: %v\n", err)
						return
					}
					if msg.Ack == nil {
						yt.SendAck(msg)
					}

				case <-yt.publisherConnected:
					node.events <- ConnectedState

				case wsErr := <-yt.wsErrors:
					log.Errorf("WS Error: %v\n", wsErr)
					return
					// if wsPinger != nil {
					//   wsPinger.Stop()
					//   wsPinger = nil
					//   wsPingerChan = nil
					// }
					//
					// _ = yt.CloseWS()
					// _ = yt.CloseRTCPublisher()
					// _ = yt.CloseRTCSubscriber()
					//
					// if err := yt.InitializeWS(); err != nil {
					//   yt.wsErrors <- fmt.Errorf("failed to initialize WS: %w", err)
					//   continue
					// }
					// yt.SendHello()

				case <-node.ctx.Done():
					log.Infof("Node cancelled\n")
					return
				}
			}
		},
	)

	return node, nil
}

func (node *Node) Stop() {
	node.cancel()
	node.wg.Wait()
}

func (node *Node) Data() <-chan []byte {
	return node.yt.rtcPacketsInc
}

func (node *Node) Send(buf []byte) {
	node.yt.rtcPacketsOut <- buf
}

func (node *Node) Events() <-chan State {
	return node.events
}

func (node *Node) IsTargetInRoom() bool {
	node.yt.peerConfigMu.RLock()
	res := node.yt.peerNames.ExistsInverse(node.yt.targetName)
	node.yt.peerConfigMu.RUnlock()

	return res
}

type State int

const (
	ConnectedState State = iota
	StoppedState
)
