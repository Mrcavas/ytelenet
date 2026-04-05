package ytnode

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
	"github.com/sirupsen/logrus"
	"github.com/vishalkuo/bimap"
)

type YTClient struct {
	log *logrus.Logger

	http *http.Client
	ws   *websocket.Conn

	chatUrl     string
	displayName string
	targetName  string

	instanceId       string
	InitData         *InitializationHTTPData
	initialStates    *StatesHTTPData
	connectedPeerIds []string

	wsMu          sync.Mutex
	wsErrors      chan error
	wsMessages    chan *WSMessageIncoming
	wsInitialized chan struct{}

	ServerHello *ServerHelloPayload

	rtcPublisher   *webrtc.PeerConnection
	publisherUfrag *string

	rtcSubscriber   *webrtc.PeerConnection
	subscriberUfrag *string

	peerNames *bimap.BiMap[PeerId, string]
	peerMids  *bimap.BiMap[PeerId, string]

	publisherConnected chan struct{}
	rtcPacketsInc      chan []byte
	rtcPacketsOut      chan []byte
}

func NewYTClient(
	log *logrus.Logger, chatUrl, displayName, targetName string,
) (yt *YTClient) {
	return &YTClient{
		log: log,

		http: nil,
		ws:   nil,

		chatUrl:     chatUrl,
		displayName: displayName,
		targetName:  targetName,

		instanceId:       uuid.NewString(),
		connectedPeerIds: make([]string, 0, 2),

		wsErrors:      make(chan error, 10),
		wsMessages:    make(chan *WSMessageIncoming, 100),
		wsInitialized: make(chan struct{}, 1),

		peerNames: bimap.NewBiMap[PeerId, string](),
		peerMids:  bimap.NewBiMap[PeerId, string](),

		publisherConnected: make(chan struct{}, 1),
		rtcPacketsInc:      make(chan []byte),
		rtcPacketsOut:      make(chan []byte),
	}
}

func (yt *YTClient) handleWSMessage(msg *WSMessageIncoming) error {
	switch {
	case msg.ServerHello != nil:
		yt.log.Debugf("Got ServerHello!")
		yt.ServerHello = msg.ServerHello
		yt.wsInitialized <- struct{}{}

	case msg.PublisherSdpAnswer != nil:
		yt.log.Debugf("Got PublisherSdpAnswer")
		err := yt.HandleRTCPublisherAnswer(msg.PublisherSdpAnswer.Sdp)
		if err != nil {
			return fmt.Errorf("failed to handle publisher answer: %w", err)
		}

	case msg.SubscriberSdpOffer != nil:
		yt.log.Debugf("Got SubscriberSdpOffer")
		err := yt.HandleRTCSubscriberOffer(msg.SubscriberSdpOffer.Sdp)
		if err != nil {
			return fmt.Errorf("failed to handle subscriber offer: %w", err)

		}

	case msg.WebrtcIceCandidate != nil:
		if msg.WebrtcIceCandidate.Target == "PUBLISHER" {
			err := yt.HandleIncomingPublisherCandidate(msg.WebrtcIceCandidate)
			if err != nil {
				yt.log.Errorf("Failed to add publisher ICE candidate: %v", err)
			}
		}

		if msg.WebrtcIceCandidate.Target == "SUBSCRIBER" {
			err := yt.HandleIncomingSubscriberCandidate(msg.WebrtcIceCandidate)
			if err != nil {
				yt.log.Errorf("Failed to add subscriber ICE candidate: %v", err)
			}
		}

	case msg.UpdateDescription != nil:
		for id := range yt.peerNames.GetForwardMap() {
			yt.peerNames.Delete(id)
		}

		for _, v := range msg.UpdateDescription.Description {
			yt.peerNames.Insert(v.Id, v.Meta.Name)
		}

	case msg.UpsertDescription != nil:
		for _, v := range msg.UpsertDescription.Description {
			yt.peerNames.Insert(v.Id, v.Meta.Name)
		}

	case msg.RemoveDescription != nil:
		for _, v := range msg.RemoveDescription.DescriptionId {
			yt.peerNames.Delete(v)
		}

	case msg.SlotsConfig != nil:
		for id := range yt.peerMids.GetForwardMap() {
			yt.peerMids.Delete(id)
		}

		for _, v := range msg.SlotsConfig.Slots {
			if v.ParticipantVideoByMid == nil {
				continue
			}
			yt.peerMids.Insert(
				v.ParticipantVideoByMid.ParticipantId, v.ParticipantVideoByMid.Mid,
			)
		}

	case msg.SelfQualityReport != nil:
		yt.log.Infof("SelfQualityReport: %v", msg.SelfQualityReport.NetworkScore)

	case msg.UpsertParticipantsQualityReport != nil:
	case msg.VadActivity != nil:
	case msg.SlotsMeta != nil:
	case msg.Ack != nil:

	default:
		yt.log.Warnf("Unknown WS message type:")
		yt.log.Warnln(prettifyJson(msg.Raw))
	}

	return nil
}

func prettifyJson(val any) string {
	pretty, err := json.MarshalIndent(val, "", "  ")
	if err != nil {
		return "unable to marshal json"
	}

	return string(pretty)
}
