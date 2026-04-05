package ytnode_man

import (
	"fmt"
	"slices"
	"sync"
	"ytelenet/ytnode"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type IndexedEvent struct {
	idx   int
	state ytnode.State
}

type NodeManager struct {
	Nodes []*ytnode.Node

	allEvents chan IndexedEvent
	allDataIn chan []byte

	mu               sync.Mutex
	connectionStates []bool
	allConnected     chan struct{}
}

func MakeNodeManager(
	amount int, log *logrus.Logger, roomUrl, name, target string,
) (*NodeManager, error) {
	if err := ytnode.InitDefaultPayloads(); err != nil {
		return nil, err
	}

	man := &NodeManager{
		Nodes: make([]*ytnode.Node, amount),

		allEvents: make(chan IndexedEvent, amount*2),
		allDataIn: make(chan []byte),

		connectionStates: make([]bool, amount),
		allConnected:     make(chan struct{}),
	}

	g := new(errgroup.Group)

	for i := 0; i < amount; i++ {
		g.Go(
			func() error {
				node, err := ytnode.MakeNew(
					log,
					roomUrl,
					fmt.Sprintf("%v|%v", name, i),
					fmt.Sprintf("%v|%v", target, i),
				)
				if err != nil {
					return err
				}
				man.Nodes[i] = node
				return nil
			},
		)
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	for i, node := range man.Nodes {
		go func() {
			for {
				select {
				case buf := <-node.Data():
					man.allDataIn <- buf

				case state := <-node.Events():
					man.allEvents <- IndexedEvent{
						idx:   i,
						state: state,
					}
					if state == ytnode.StoppedState {
						return
					}

					man.mu.Lock()
					man.connectionStates[i] = true
					if all(man.connectionStates) {
						man.allConnected <- struct{}{}
					}
					man.mu.Unlock()
				}
			}
		}()
	}

	return man, nil
}

func (man *NodeManager) SendTo(idx int, buf []byte) {
	man.Nodes[idx].Send(buf)
}

func (man *NodeManager) SendAll(buf []byte) {
	for i, node := range man.Nodes {
		node.Send(slices.Concat([]byte(fmt.Sprintf("%v | ", i)), buf))
	}
}

func (man *NodeManager) Data() <-chan []byte {
	return man.allDataIn
}

func (man *NodeManager) Events() <-chan IndexedEvent {
	return man.allEvents
}

func (man *NodeManager) AllConnected() <-chan struct{} {
	return man.allConnected
}

func (man *NodeManager) Stop() {
	var wg sync.WaitGroup
	for _, node := range man.Nodes {
		wg.Go(node.Stop)
	}
	wg.Wait()
}

func all(slice []bool) bool {
	for _, val := range slice {
		if !val {
			return false
		}
	}
	return true
}
