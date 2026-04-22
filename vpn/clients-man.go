package vpn

import (
	"bytes"
	"fmt"
	"sync"
	"time"
	"ytelenet/ytnode"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type NodeManager struct {
	internalLog *log.Logger
	db          *ClientsDB
	nodes       []*ytnode.Node
	pcNumToIdx  map[byte]int

	allDataIn chan []byte

	stopped bool
}

func MakeClientsManager(
	internalLog *log.Logger, db *ClientsDB,
) (*NodeManager, error) {
	amount := len(db.Clients)

	man := &NodeManager{
		internalLog: internalLog,
		db:          db,
		nodes:       make([]*ytnode.Node, amount),

		pcNumToIdx: make(map[byte]int, amount),

		allDataIn: make(chan []byte, 8192),
	}

	g := new(errgroup.Group)

	for i, client := range db.Clients {
		man.pcNumToIdx[client.PcNum] = i

		g.Go(
			func() error {
				return makeNode(i, man, internalLog, client)
			},
		)
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	for i, client := range db.Clients {
		go man.handleNode(i, client)
	}

	return man, nil
}

func makeNode(
	idx int, man *NodeManager, internalLog *log.Logger, client ClientData,
) error {
	node, err := ytnode.MakeNew(
		internalLog,
		makeRoomUrl(client.RoomId),
		"server",
		client.Name,
	)
	if err != nil {
		return err
	}

	st := <-node.Events()
	if st != ytnode.ConnectedState {
		return fmt.Errorf("unable to connect to node %v", client.Name)
	}

	man.nodes[idx] = node
	return nil
}

func (man *NodeManager) handleNode(idx int, client ClientData) {
	node := man.nodes[idx]

	for {
		select {
		case buf := <-node.Data():
			man.allDataIn <- bytes.Clone(buf)

		case state := <-node.Events():
			if state == ytnode.StoppedState {
				man.nodes[idx] = nil
				go man.TryReconnect(idx, client, 0)
				return
			}
		}
	}
}

func (man *NodeManager) TryReconnect(idx int, client ClientData, tryIdx int) {
	if man.stopped {
		return
	}

	// if tryIdx == 3 {
	//   log.Errorln("Reconnection failed")
	//   man.Stop()
	//   return
	// }

	<-time.After( /*time.Duration(tryIdx) * */ 2 * time.Second)

	err := makeNode(idx, man, man.internalLog, client)
	if err != nil {
		man.TryReconnect(idx, client, tryIdx+1)
		return
	}

	go man.handleNode(idx, client)
}

func (man *NodeManager) SendTo(pcNum byte, buf []byte) {
	idx, ok := man.pcNumToIdx[pcNum]
	if !ok {
		log.Debugf("Failed to get node index for pcNum %v", pcNum)
		return
	}

	if man.nodes[idx] != nil {
		man.nodes[idx].Send(buf)
	}
}

func (man *NodeManager) Data() <-chan []byte {
	return man.allDataIn
}

func (man *NodeManager) Stop() {
	man.stopped = true

	var wg sync.WaitGroup
	for _, node := range man.nodes {
		if node != nil {
			wg.Go(node.Stop)
		}
	}
	wg.Wait()
}
