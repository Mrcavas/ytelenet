package vpn

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
	"ytelenet/ytnode"

	log "github.com/sirupsen/logrus"
)

// The desktop GUI can't bring up the tunnel itself, since that needs admin
// rights, so it launches this mode elevated and the backend connects back to
// it over localhost.
//
// The protocol is JSON lines. The backend sends "hello" with the key it was
// launched with, and the GUI answers with "start" holding the client options.
// Then the backend streams "log" entries, "connected" once the tunnel is up and
// the final "stopped" (with the error, if it failed). The GUI stops the client
// by sending "stop" or just closing the connection.
type backendMessage struct {
	Type string `json:"type"`

	// hello
	Key string `json:"key,omitempty"`

	// start
	Token       string `json:"token,omitempty"`
	Dns         string `json:"dns,omitempty"`
	NoAutoRoute bool   `json:"noAutoRoute,omitempty"`
	Debug       bool   `json:"debug,omitempty"`

	// log
	Level string `json:"level,omitempty"`
	Msg   string `json:"msg,omitempty"`
	Time  string `json:"time,omitempty"`

	// stopped
	Error string `json:"error,omitempty"`
}

type backendConn struct {
	mu  sync.Mutex
	enc *json.Encoder
}

func (c *backendConn) send(msg backendMessage) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.enc.Encode(msg)
}

// Write forwards the entries of logrus' JSONFormatter to the GUI
func (c *backendConn) Write(p []byte) (int, error) {
	entry := Log{}
	if err := json.Unmarshal(p, &entry); err != nil {
		return 0, err
	}

	err := c.send(backendMessage{
		Type: "log", Level: entry.Level, Msg: entry.Msg, Time: entry.Time,
	})
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func BackendMain(
	interrupt chan os.Signal, addr, key string, mtu uint32,
	batchingOpts ytnode.BatchingOptions,
) error {
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to the GUI: %w", err)
	}
	defer conn.Close()

	out := &backendConn{enc: json.NewEncoder(conn)}
	if err := out.send(backendMessage{Type: "hello", Key: key}); err != nil {
		return fmt.Errorf("failed to send hello: %w", err)
	}

	dec := json.NewDecoder(conn)
	var start backendMessage
	if err := dec.Decode(&start); err != nil {
		return fmt.Errorf("failed to read the start message: %w", err)
	}
	if start.Type != "start" {
		return fmt.Errorf("expected the start message, got %q", start.Type)
	}

	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(out)
	defer log.SetOutput(os.Stderr)
	if start.Debug {
		log.SetLevel(log.DebugLevel)
	}

	stop := make(chan struct{})
	var stopOnce sync.Once
	stopClient := func() { stopOnce.Do(func() { close(stop) }) }

	go func() {
		defer stopClient()
		for {
			var msg backendMessage
			// an error means the GUI closed the connection (or died)
			if err := dec.Decode(&msg); err != nil || msg.Type == "stop" {
				return
			}
		}
	}()
	go func() {
		<-interrupt
		stopClient()
	}()

	err = runClient(
		stop, start.Token, start.Debug, TunnelOptions{
			MTU:         mtu,
			Target:      "server",
			NoAutoRoute: start.NoAutoRoute,
			Dns:         start.Dns,
		}, batchingOpts, func() {
			_ = out.send(backendMessage{Type: "connected"})
		},
	)

	stopped := backendMessage{Type: "stopped"}
	if err != nil {
		stopped.Error = err.Error()
	}
	_ = out.send(stopped)

	return err
}
