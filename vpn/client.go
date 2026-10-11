package vpn

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"net/netip"
	"os"
	"strconv"
	"strings"
	"ytelenet/ytnode"

	log "github.com/sirupsen/logrus"
)

func ClientMain(
	interrupt chan os.Signal, initStr string, debug bool, opts TunnelOptions,
	batchingOpts ytnode.BatchingOptions,
) {
	stop := make(chan struct{})
	go func() {
		<-interrupt
		close(stop)
	}()

	if err := runClient(stop, initStr, debug, opts, batchingOpts, nil); err != nil {
		log.Fatalf("%v\n", err)
	}
}

// parseToken decodes base64("<room id>;<client name>;<pc number>")
func parseToken(initStr string) (roomUrl, clientName string, pcNum int, err error) {
	token, err := base64.StdEncoding.DecodeString(initStr)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to decode token")
	}

	parts := strings.Split(string(token), ";")
	if len(parts) != 3 {
		return "", "", 0, fmt.Errorf("failed to decode token")
	}

	pcNum, err = strconv.Atoi(parts[2])
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to decode token")
	}

	return makeRoomUrl(parts[0]), parts[1], pcNum, nil
}

// runClient connects to the room and forwards packets between it and the
// tunnel until stop is closed or the node stops. onConnected (if not nil) is
// called once the tunnel is up.
func runClient(
	stop <-chan struct{}, initStr string, debug bool, opts TunnelOptions,
	batchingOpts ytnode.BatchingOptions, onConnected func(),
) error {
	log.Infof("Launching client\n")

	internalLog := makeInternalLog(debug)
	cleanupLeftoverDNS()

	roomUrl, clientName, pcNum, err := parseToken(initStr)
	if err != nil {
		return err
	}

	log.Infof("Initializing YT node\n")
	node, err := ytnode.MakeNew(
		internalLog, roomUrl, clientName, opts.Target, batchingOpts,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize YT nodes: %w", err)
	}
	defer func() { node.Stop() }()

	log.Infof("Waiting for connection")
	select {
	case st := <-node.Events():
		if st != ytnode.ConnectedState {
			return fmt.Errorf("unable to connect")
		}
	case <-stop:
		log.Info("Interrupted\n")
		return nil
	}
	log.Infof("Connected\n")

	if opts.Target == "server" {
		if !node.IsTargetInRoom() {
			return fmt.Errorf("server isn't connected, try reconnecting")
		}
		log.Infof("Server is in the room, all good")
	}

	tunnel, err := makeAndStartTunnel(internalLog, true, pcNum, opts, nil)
	if err != nil {
		return err
	}
	defer tunnel.Close()

	if !opts.NoAutoRoute && opts.Dns != "" {
		restoreDNS, err := setSystemDNS(tunnelName, netip.MustParseAddr(opts.Dns))
		if err != nil {
			log.Errorf("Failed to set the system DNS: %v", err)
		} else {
			defer restoreDNS()
		}
	}

	if onConnected != nil {
		onConnected()
	}

	tunnelErr := make(chan error, 1)
	go func() {
		buf := make([]byte, opts.MTU)

		for {
			size, err := tunnel.Read(buf)
			if errors.Is(err, os.ErrClosed) {
				log.Infof("Closed tunnel\n")
				return
			}
			if err != nil {
				tunnelErr <- fmt.Errorf("failed to read from tunnel: %w", err)
				return
			}

			if size < 20 {
				continue
			}

			log.Debugf("sending ip: %v", buf[:size])
			node.Send(bytes.Clone(buf[:size]))
		}
	}()

	for {
		select {
		case state := <-node.Events():
			if state == ytnode.StoppedState {
				log.Warnf("Node stopped\n")
				return nil
			}

		case err := <-tunnelErr:
			return err

		case buf := <-node.Data():
			if len(buf) < 20 {
				break
			}
			_, err := tunnel.Write(bytes.Clone(buf))
			if errors.Is(err, os.ErrClosed) {
				return nil
			}
			if err != nil {
				log.Errorf("Couldn't write packet\n")
			}

		case <-stop:
			log.Infof("Interrupted\n")
			return nil
		}
	}
}
