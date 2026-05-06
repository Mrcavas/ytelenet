package vpn

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"ytelenet/ytnode"

	log "github.com/sirupsen/logrus"
)

var stopChan chan struct{}

type Log struct {
	Level string `json:"level"`
	Msg   string `json:"msg"`
	Time  string `json:"time"`
}

type AndroidVoidFunction interface {
	Exec()
}
type AndroidStringFunction interface {
	Exec(data string)
}
type AndroidLogFunction interface {
	Exec(data *Log)
}

type androidLogWriter struct {
	onLog AndroidLogFunction
}

func (wr androidLogWriter) Write(p []byte) (int, error) {
	logEntry := Log{}
	if err := json.Unmarshal(p, &logEntry); err != nil {
		return 0, err
	}
	wr.onLog.Exec(&logEntry)
	return len(p), nil
}

//goland:noinspection GoUnusedExportedFunction
func Start(
	fd int, token string, onLog AndroidLogFunction,
	onConnected AndroidVoidFunction,
	onStop AndroidStringFunction,
) {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(androidLogWriter{onLog: onLog})
	log.SetLevel(log.DebugLevel)

	stopChan = make(chan struct{}, 1)

	go func() {
		err := androidMain(
			stopChan, token, &TunnelOptions{
				Dns: "8.8.8.8",
			}, fd, onConnected.Exec,
		)
		if err != nil {
			onStop.Exec(err.Error())
		} else {
			onStop.Exec("")
		}
	}()
}

//goland:noinspection GoUnusedExportedFunction
func Stop() {
	log.Debugf("stopChan <- struct{}{}")

	select {
	case <-stopChan:
	default:
	}
	stopChan <- struct{}{}
}

func androidMain(
	interrupt chan struct{}, initStr string, opts *TunnelOptions, fd int,
	onConnected func(),
) error {
	log.Infof("Launching client\n")

	internalLog := makeInternalLog(false)

	token, err := base64.StdEncoding.DecodeString(initStr)
	if err != nil {
		return fmt.Errorf("failed to decode token")
	}

	parts := strings.Split(string(token), ";")
	if len(parts) != 3 {
		return fmt.Errorf("failed to decode token")
	}

	roomUrl := makeRoomUrl(parts[0])
	clientName := parts[1]
	pcNum, err := strconv.Atoi(parts[2])
	if err != nil {
		return fmt.Errorf("failed to decode token")
	}

	tunnel := makeAndStartTunnel(internalLog, true, pcNum, opts, &fd)
	defer tunnel.Close()

	log.Infof("Initializing YT node\n")
	node, err := ytnode.MakeNew(internalLog, roomUrl, clientName, "server")
	if err != nil {
		return fmt.Errorf("failed to initialize YT nodes: %w\n", err)
	}
	defer node.Stop()

	log.Infof("Waiting for connection")
	select {
	case st := <-node.Events():
		if st != ytnode.ConnectedState {
			return fmt.Errorf("unable to connect")
		}
	case <-interrupt:
		log.Info("Interrupted\n")
		return nil
	}

	if node.IsTargetInRoom() {
		log.Infof("Connected to server")
		onConnected()
	} else {
		log.Errorf("Server isn't connected. Try reconnecting")
		return nil
	}

	go func() {
		buf := make([]byte, opts.MTU)

		for {
			size, err := tunnel.Read(buf)
			if errors.Is(err, os.ErrClosed) {
				log.Infof("Closed tunnel\n")
				break
			}
			if err != nil {
				log.Fatalf("Failed to read from tunnel: %v\n", err)
			}

			node.Send(bytes.Clone(buf[:size]))
		}
	}()

	for {
		select {
		case state := <-node.Events():
			if state == ytnode.StoppedState {
				log.Infof("Node stopped\n")
				return nil
			}

		case buf := <-node.Data():
			_, err := tunnel.Write(bytes.Clone(buf))
			if errors.Is(err, os.ErrClosed) {
				return nil
			}
			if err != nil {
				log.Errorf("Couldn't write packet: %v\n", err)
			}

		case <-interrupt:
			log.Infof("Interrupted\n")
			return nil
		}
	}
}
