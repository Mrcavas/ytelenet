package vpn

import (
	"encoding/json"
	"os"
)

type ClientData struct {
	RoomId string `json:"roomId"`
	PcNum  byte   `json:"pcNum"`
	Name   string `json:"name"`
}

type ClientsDB struct {
	Clients []ClientData `json:"clients"`
}

func ParseClients() (*ClientsDB, error) {
	f, err := os.Open("clients.json")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var clients ClientsDB
	if err := json.NewDecoder(f).Decode(&clients); err != nil {
		return nil, err
	}

	return &clients, nil
}
