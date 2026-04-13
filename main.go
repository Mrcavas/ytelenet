package main

import (
	// "fmt"
	"os"
	"os/signal"
	"ytelenet/vpn"

	"github.com/alexflint/go-arg"
	log "github.com/sirupsen/logrus"
)

type ClientCmd struct {
	Token       string `arg:"positional,required"`
	Dns         string `arg:"--dns" default:"8.8.8.8"`
	NoAutoRoute bool   `arg:"--no-auto-route"`
}
type ServerCmd struct{}

type __args__ struct {
	Client *ClientCmd `arg:"subcommand:client"`
	Server *ServerCmd `arg:"subcommand:server"`
}

func (__args__) Version() string {
	return "YTelenet 0.5.0"
}

var args __args__

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	log.SetFormatter(
		&log.TextFormatter{
			FullTimestamp:   true,
			ForceColors:     true,
			TimestampFormat: "15:04:05.000",
		},
	)

	p := arg.MustParse(&args)
	if p.Subcommand() == nil {
		p.Fail("Must select working mode")
	}

	if args.Server != nil {
		clients, err := vpn.ParseClients()
		if err != nil {
			log.Fatalf("Failed to parse clients.json: %v", err)
		}

		vpn.ServerMain(interrupt, clients)
	} else if args.Client != nil {
		vpn.ClientMain(
			interrupt, args.Client.Token, &vpn.TunnelOptions{
				NoAutoRoute: args.Client.NoAutoRoute,
				Dns:         args.Client.Dns,
			},
		)
	}
}
