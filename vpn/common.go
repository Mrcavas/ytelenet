package vpn

import (
	"fmt"
	"net/netip"

	tun "github.com/sagernet/sing-tun"
	log "github.com/sirupsen/logrus"
)

func makeInternalLog() *log.Logger {
	internalLog := log.New()
	internalLog.SetFormatter(log.StandardLogger().Formatter)
	internalLog.SetLevel(log.WarnLevel)

	return internalLog
}

type TunnelOptions struct {
	NoAutoRoute bool
	Dns         string
}

func makeAndStartTunnel(
	internalLog *log.Logger, isClient bool, pcNum int, opts *TunnelOptions,
	fd *int,
) tun.Tun {
	netMon, err := tun.NewNetworkUpdateMonitor(internalLog)
	if err != nil {
		log.Fatalf("Failed to create network update monitor (???): %v\n", err)
	}

	interfaceMon, err := tun.NewDefaultInterfaceMonitor(
		netMon,
		internalLog,
		tun.DefaultInterfaceMonitorOptions{},
	)
	if err != nil {
		log.Fatalf("Failed to create interface monitor (???): %v\n", err)
	}

	mask := 24
	if isClient {
		mask = 32
	}

	tunnelOpts := tun.Options{
		MTU:              1186,
		InterfaceMonitor: interfaceMon,
	}

	if fd == nil {
		tunnelOpts.Name = "YTelenet"
		tunnelOpts.AutoRoute = isClient && !opts.NoAutoRoute
		tunnelOpts.StrictRoute = isClient
		tunnelOpts.Inet4Address = []netip.Prefix{
			netip.MustParsePrefix(fmt.Sprintf("42.42.42.%v/%v", pcNum, mask)),
		}
	} else {
		tunnelOpts.FileDescriptor = *fd
	}

	if isClient {
		dnsAddr, err := netip.ParseAddr(opts.Dns)
		if err != nil {
			log.Fatalf("Failed to parse DNS")
		}

		tunnelOpts.DNSServers = []netip.Addr{dnsAddr}
	}

	tunnel, err := tun.New(tunnelOpts)
	if err != nil {
		log.Fatalf("Failed to create tunnel: %v\n", err)
	}

	if err := tunnel.Start(); err != nil {
		log.Fatalf("Failed to start tunnel: %v\n", err)
	}
	log.Infof("Started tunnel\n")

	return tunnel
}

func makeRoomUrl(roomId string) string {
	return fmt.Sprintf("https://telemost.yandex.ru/j/%v", roomId)
}
