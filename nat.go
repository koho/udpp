package udpp

import (
	"fmt"
	"net"

	"github.com/fatedier/frp/pkg/nathole"
)

type NAT struct {
	LocalAddr   *net.UDPAddr
	RemoteAddrs []*net.UDPAddr
	Feature     *nathole.NatFeature
}

func Discover(server string) (*NAT, error) {
	addrs, localAddr, err := nathole.Discover([]string{server}, "")
	if err != nil {
		return nil, err
	}
	if len(addrs) < 2 {
		return nil, fmt.Errorf("can not get enough addresses, need 2, got: %v\n", addrs)
	}
	localIPs, _ := nathole.ListLocalIPsForNatHole(10)

	natFeature, err := nathole.ClassifyNATFeature(addrs, localIPs)
	if err != nil {
		return nil, err
	}
	remoteAddrs := make([]*net.UDPAddr, 0, len(addrs))
	for _, addr := range addrs {
		if a, err := net.ResolveUDPAddr("udp", addr); err == nil {
			remoteAddrs = append(remoteAddrs, a)
		}
	}
	return &NAT{
		LocalAddr:   localAddr.(*net.UDPAddr),
		RemoteAddrs: remoteAddrs,
		Feature:     natFeature,
	}, nil
}
