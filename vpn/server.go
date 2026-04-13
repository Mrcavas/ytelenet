package vpn

import (
	"errors"
	"os"

	log "github.com/sirupsen/logrus"
)

func ServerMain(interrupt chan os.Signal, clients *ClientsDB) {
	log.Infof("Launching server\n")

	internalLog := makeInternalLog()

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

			nodes.SendTo(buf[19], buf[:size])
		}
	}()

	for {
		select {
		case buf := <-nodes.Data():
			_, err := tunnel.Write(buf)
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
