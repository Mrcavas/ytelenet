//go:build !linux

package vpn

import "net/netip"

// setSystemDNS is a no-op where sing-tun configures DNS itself
func setSystemDNS(string, netip.Addr) (func(), error) {
	return func() {}, nil
}

func cleanupLeftoverDNS() {}
