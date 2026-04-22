package ytnode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
)

const DefaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:147.0) Gecko/20100101 Firefox/147.0"

func (yt *YTClient) Initialize() error {
	yt.log.Debugf("Initializing HTTP\n")

	yt.http = &http.Client{
		Timeout: 5 * time.Second,
	}

	reqUrl := fmt.Sprintf(
		"https://cloud-api.yandex.ru/telemost_front/v2/telemost"+
			"/conferences/%v"+
			"/connection?next_gen_media_platform_allowed=true"+
			"&display_name=%v&waiting_room_supported=true",
		url.QueryEscape(yt.chatUrl), url.QueryEscape(yt.displayName),
	)

	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Host", "cloud-api.yandex.ru")
	req.Header.Set("User-Agent", DefaultUserAgent)
	req.Header.Set("X-Telemost-Client-Version", "179.3.0")
	req.Header.Set("idempotency-key", uuid.NewString())
	req.Header.Set("Origin", "https://telemost.yandex.ru")
	req.Header.Set("Referer", "https://telemost.yandex.ru/")
	req.Header.Set("Client-Instance-Id", yt.instanceId)

	resp, err := yt.http.Do(req)
	if err != nil {
		return fmt.Errorf("network error: %w", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status code: %v", resp.StatusCode)
	}

	yt.InitData = &InitializationHTTPData{}
	if err := json.NewDecoder(resp.Body).Decode(yt.InitData); err != nil {
		return fmt.Errorf("invalid json: %w", err)
	}

	return nil
}

func (yt *YTClient) RequestStatesForNewPeer(peerId string) (
	*StatesHTTPData, error,
) {
	return yt.fetchStates(
		[]PeerWithVersion{
			{
				PeerID: peerId,
			},
		},
	)
}

func (yt *YTClient) PingStates() (*StatesHTTPData, error) {
	yt.log.Debugf("HTTP ping!\n")

	if yt.InitData == nil {
		return nil, fmt.Errorf("PingStates requires YTClient to be initialized")
	}

	peers := make([]PeerWithVersion, len(yt.connectedPeerIds)+1)
	zeroVersion := 0

	peers[0] = PeerWithVersion{
		PeerID:  yt.InitData.PeerID,
		Version: &zeroVersion,
	}

	for i, peerId := range yt.connectedPeerIds {
		peers[i+1] = PeerWithVersion{
			PeerID:  peerId,
			Version: &zeroVersion,
		}
	}

	return yt.fetchStates(peers)
}

func (yt *YTClient) fetchStates(peers []PeerWithVersion) (
	*StatesHTTPData, error,
) {
	reqUrl := fmt.Sprintf(
		"https://cloud-api.yandex.ru/telemost_front/v2/telemost/conferences/%v/request-states",
		url.QueryEscape(yt.chatUrl),
	)

	var permissionsVersion *int
	conferenceVersion := -1
	if yt.initialStates != nil {
		permissionsVersion = &yt.initialStates.Permissions.Version
		conferenceVersion = yt.initialStates.Conference.Version
	}

	reqBody := StatesHTTPRequestData{
		Peers: peers,
		Conference: ConferenceState{
			Version: conferenceVersion,
		},
		Permissions: PermissionsState{
			Version: permissionsVersion,
		},
	}

	reqBodyJson, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request body: %w", err)
	}

	req, err := http.NewRequest("POST", reqUrl, bytes.NewReader(reqBodyJson))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Host", "cloud-api.yandex.ru")
	req.Header.Set("User-Agent", DefaultUserAgent)
	req.Header.Set("idempotency-key", uuid.NewString())
	req.Header.Set("Origin", "https://telemost.yandex.ru")
	req.Header.Set("Referer", "https://telemost.yandex.ru/")
	req.Header.Set("Client-Instance-Id", yt.instanceId)

	resp, err := yt.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("network error: %w", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %v", resp.StatusCode)
	}

	states := &StatesHTTPData{}
	if err := json.NewDecoder(resp.Body).Decode(states); err != nil {
		return nil, fmt.Errorf("invalid json: %w", err)
	}

	if yt.initialStates == nil {
		yt.initialStates = states
	}

	return states, nil
}
