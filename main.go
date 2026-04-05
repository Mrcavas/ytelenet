package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"ytelenet/ytnode_man"

	"github.com/alexflint/go-arg"
	log "github.com/sirupsen/logrus"
)

const (
	chatUrl = "https://telemost.yandex.ru/j/91598352795911"
	amount  = 4
)

var scanner *bufio.Scanner

type __args__ struct {
	// Amount int    `arg:"positional,required" help:"Amount of nodes to spin up"`
	Name   string `arg:"positional,required" help:"This client name"`
	Target string `arg:"positional,required" help:"Target client name"`
}

func (__args__) Version() string {
	return "YTClient 0.0.0"
}

var args __args__

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	scanner = bufio.NewScanner(os.Stdin)

	log.SetFormatter(
		&log.TextFormatter{
			FullTimestamp:   true,
			ForceColors:     true,
			TimestampFormat: "15:04:05.000",
		},
	)
	// log.SetLevel(log.DebugLevel)

	arg.MustParse(&args)
	log.Infof(
		"Starting %v nodes '%v' to '%v'\n", amount, args.Name, args.Target,
	)

	voidLog := log.New()
	voidLog.SetOutput(io.Discard)

	nodes, err := ytnode_man.MakeNodeManager(
		amount, voidLog, chatUrl, args.Name, args.Target,
	)
	if err != nil {
		log.Fatalln(err)
	}

	<-nodes.AllConnected()
	go scanInput(nodes)
	log.Infof("Connected\n")

	for {
		select {
		case buf := <-nodes.Data():
			handleRTCPacket(buf)

		case <-interrupt:
			log.Infof("Interrupted\n")
			nodes.Stop()
			return
		}
	}
}

func handleRTCPacket(buf []byte) {
	if len(buf) >= 1024 && len(buf) <= 1030 && buf[8] == 97 {
		return
	}
	if len(buf) >= 1024 && buf[8] == 97 {
		log.Infof("[%v]: *large packet with size %v*", args.Target, len(buf))
		return
	}
	log.Infof("[%v]: %s", args.Target, buf)
}

func scanInput(nodes *ytnode_man.NodeManager) {
	for scanner.Scan() {
		text := scanner.Text()

		switch {
		case strings.HasPrefix(text, "/burst "):
			amount, err := strconv.Atoi(text[7:])
			if err != nil {
				return
			}
			filledBuf := make([]byte, 1024)
			for i := 1; i < 1024; i++ {
				filledBuf[i] = 97
			}

			log.Infof("[%v]: %v packet burst start\n", args.Name, amount)
			nodes.SendAll([]byte(fmt.Sprintf("%v packet burst start", amount)))
			for i := 1; i <= amount; i++ {
				nodes.SendAll(filledBuf)
			}
			log.Infof("[%v]: %v packet burst end\n", args.Name, amount)
			nodes.SendAll([]byte(fmt.Sprintf("%v packet burst end", amount)))

		case strings.HasPrefix(text, "/kb "):
			kb, err := strconv.Atoi(text[4:])
			if err != nil {
				return
			}

			amount := kb * 1024
			filledBuf := make([]byte, amount)
			for i := 1; i < amount; i++ {
				filledBuf[i] = 97
			}

			log.Infof("[%v]: *large packet with size %v* \n", args.Name, amount)
			nodes.SendAll(filledBuf)

		default:
			log.Infof("[%v]: %v\n", args.Name, text)
			nodes.SendAll([]byte(text))
		}
	}
}
