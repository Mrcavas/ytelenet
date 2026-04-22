package vpn

import (
	"encoding/base64"
	"errors"
	"os"
	"strconv"
	"strings"
	"ytelenet/ytnode"

	log "github.com/sirupsen/logrus"
)

func ClientMain(
	interrupt chan os.Signal, initStr string, opts *TunnelOptions,
) {
	log.Infof("Launching client\n")

	internalLog := makeInternalLog(false)

	token, err := base64.StdEncoding.DecodeString(initStr)
	if err != nil {
		log.Fatalf("Failed to decode token")
	}

	parts := strings.Split(string(token), ";")
	if len(parts) != 3 {
		log.Fatalf("Failed to decode token")
	}

	roomUrl := makeRoomUrl(parts[0])
	clientName := parts[1]
	pcNum, err := strconv.Atoi(parts[2])
	if err != nil {
		log.Fatalf("Failed to decode token")
	}

	log.Infof("Initializing YT node\n")
	node, err := ytnode.MakeNew(
		internalLog, roomUrl, clientName, opts.Destination,
	)
	if err != nil {
		log.Fatalf("Failed to initialize YT nodes: %v\n", err)
	}
	defer func() { node.Stop() }()

	log.Infof("Waiting for connection")
	select {
	case st := <-node.Events():
		if st != ytnode.ConnectedState {
			log.Fatalf("Unable to connect")
		}
	case <-interrupt:
		log.Info("Interrupted\n")
		return
	}
	log.Infof("Connected\n")

	tunnel := makeAndStartTunnel(internalLog, true, pcNum, opts, nil)
	defer tunnel.Close()

	go func() {
		buf := make([]byte, 1186)

		for {
			size, err := tunnel.Read(buf)
			if errors.Is(err, os.ErrClosed) {
				log.Infof("Closed tunnel\n")
				break
			}
			if err != nil {
				log.Fatalf("Failed to read from tunnel: %v\n", err)
			}

			node.Send(buf[:size])
		}
	}()

	for {
		select {
		case state := <-node.Events():
			if state == ytnode.StoppedState {
				log.Warnf("Node stopped\n")
				return
			}

		case buf := <-node.Data():
			_, err := tunnel.Write(buf)
			if errors.Is(err, os.ErrClosed) {
				return
			}
			if err != nil {
				log.Errorf("Couldn't write packet\n")
			}

		case <-interrupt:
			log.Infof("Interrupted\n")
			return
		}
	}
}
