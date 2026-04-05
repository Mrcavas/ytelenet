package ytnode

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func (yt *YTClient) InitializeWS() error {
	yt.log.Debugf("Initializing WS\n")

	if yt.InitData == nil {
		return fmt.Errorf("InitializeWS requires YTClient to be initialized first")
	}

	headers := make(http.Header)
	headers.Set("Origin", "https://telemost.yandex.ru")
	headers.Set("User-Agent", DefaultUserAgent)

	wsUrl := yt.InitData.ClientConfiguration.MediaServerURL
	ws, _, err := websocket.DefaultDialer.Dial(wsUrl, headers)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}

	yt.ws = ws

	go func() {
		defer close(yt.wsMessages)
		for {
			raw := json.RawMessage{}
			if err := ws.ReadJSON(&raw); err != nil {
				yt.wsErrors <- err
				return
			}

			msg := &WSMessageIncoming{}
			if err := json.Unmarshal(raw, msg); err != nil {
				yt.wsErrors <- err
				return
			}
			msg.Raw = &raw
			yt.wsMessages <- msg
		}
	}()

	return nil
}

func (yt *YTClient) CloseWS() {
	defer yt.ws.Close()

	closeMsg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
	err := yt.ws.WriteMessage(websocket.CloseMessage, closeMsg)
	if err != nil {
		yt.wsErrors <- fmt.Errorf("failed to send close message: %w", err)
	}
}

func (yt *YTClient) SendWS(msg *WSMessageOutgoing) {
	msg.Uid = uuid.NewString()
	yt.sendWSRaw(msg)
}

func (yt *YTClient) sendWSRaw(msg *WSMessageOutgoing) {
	yt.wsMu.Lock()
	defer yt.wsMu.Unlock()

	if msg.Ack == nil {
		yt.log.Debugf("Sending WS message:\n")
		yt.log.Debugln(prettifyJson(msg))
	}

	if err := yt.ws.WriteJSON(msg); err != nil {
		yt.wsErrors <- err
	}
}

func (yt *YTClient) SendAck(inc *WSMessageIncoming) {
	okAckPayload, err := GetDefaultPayload[AckPayload]("okAck")
	if err != nil {
		yt.wsErrors <- fmt.Errorf("failed to get okAck payload: %w", err)
		return
	}

	yt.sendWSRaw(
		&WSMessageOutgoing{
			Uid: inc.Uid,
			Ack: okAckPayload,
		},
	)
}

func (yt *YTClient) SendPing() {
	yt.SendWS(
		&WSMessageOutgoing{
			Ping: &struct{}{},
		},
	)
}

func (yt *YTClient) SendHello() {
	helloPayload, err := GetDefaultPayload[HelloPayload]("hello")
	if err != nil {
		yt.wsErrors <- fmt.Errorf("failed to get hello payload: %w", err)
		return
	}

	if yt.InitData == nil {
		yt.wsErrors <- fmt.Errorf("failed to send hello: InitData is nil")
		return
	}

	helloPayload.ParticipantMeta.Name = yt.displayName
	helloPayload.ParticipantAttributes.Name = yt.displayName
	helloPayload.ParticipantID = yt.InitData.PeerID
	helloPayload.RoomID = yt.InitData.RoomID
	helloPayload.Credentials = yt.InitData.Credentials
	helloPayload.SdkInitializationID = uuid.NewString()
	helloPayload.SdkInfo.UserAgent = DefaultUserAgent

	yt.SendWS(
		&WSMessageOutgoing{
			Hello: helloPayload,
		},
	)

	yt.log.Debugf("Sent Hello\n")
}
