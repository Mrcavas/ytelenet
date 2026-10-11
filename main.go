package main

import (
	// "fmt"
	"os"
	"os/signal"
	"syscall"
	"ytelenet/vpn"
	"ytelenet/ytnode"

	"github.com/alexflint/go-arg"
	log "github.com/sirupsen/logrus"
)

type ClientCmd struct {
	Token       string `arg:"positional,required"`
	Dns         string `arg:"--dns" default:"1.1.1.1"`
	NoAutoRoute bool   `arg:"--no-auto-route"`
	Target      string `arg:"--to" default:"server"`
}
type ServerCmd struct{}

type __args__ struct {
	Client *ClientCmd `arg:"subcommand:client"`
	Server *ServerCmd `arg:"subcommand:server"`

	Debug bool   `arg:"--debug" default:"false"`
	MTU   uint32 `arg:"--mtu" help:"tunnel MTU, by default an IP packet fills exactly two RTP packets"`

	MaxBatchSize int `arg:"-b,--batch" default:"64" help:"max RTP packets per video frame"`
}

func (__args__) Version() string {
	return "YTelenet 0.7.1"
}

var args = __args__{MTU: ytnode.DefaultTunnelMTU}

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

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

	batchingOpts := ytnode.BatchingOptions{
		MaxBatchSize: args.MaxBatchSize,
	}

	if args.Server != nil {
		clients, err := vpn.ParseClients()
		if err != nil {
			log.Fatalf("Failed to parse clients.json: %v", err)
		}

		vpn.ServerMain(interrupt, clients, args.MTU, args.Debug, batchingOpts)
	} else if args.Client != nil {
		vpn.ClientMain(
			interrupt,
			args.Client.Token,
			args.Debug,
			vpn.TunnelOptions{
				MTU:         args.MTU,
				Target:      args.Client.Target,
				NoAutoRoute: args.Client.NoAutoRoute,
				Dns:         args.Client.Dns,
			},
			batchingOpts,
		)
	}
}
