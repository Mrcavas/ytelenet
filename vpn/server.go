package vpn

import (
	"bytes"
	"errors"
	"os"

	log "github.com/sirupsen/logrus"
)

func ServerMain(interrupt chan os.Signal, clients *ClientsDB, debug bool) {
	log.Infof("Launching server\n")

	internalLog := makeInternalLog(debug)

	log.Infof("Connecting to YT\n")
	nodes, err := MakeClientsManager(internalLog, clients)
	if err != nil {
		log.Fatalf("Failed to initialize YT nodes: %v\n", err)
	}
	defer nodes.Stop()

	tunnel := makeAndStartTunnel(internalLog, false, 1, nil, nil)
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

			if size < 20 {
				continue
			}

			nodes.SendTo(buf[19], bytes.Clone(buf[:size]))
		}
	}()

	for {
		select {
		case buf := <-nodes.Data():
			if len(buf) < 20 {
				break
			}
			if buf[16] == 42 && buf[17] == 42 && buf[18] == 42 {
				idx, ok := nodes.pcNumToIdx[buf[19]]
				if !ok {
					goto writeToTun
				}
				nodes.nodes[idx].Send(buf)
				break
			}

		writeToTun:
			_, err := tunnel.Write(bytes.Clone(buf))
			if errors.Is(err, os.ErrClosed) {
				break
			}
			if err != nil {
				log.Errorf("Couldn't write packet")
			}

		case <-interrupt:
			log.Infof("Interrupted\n")
			return
		}
	}
}
