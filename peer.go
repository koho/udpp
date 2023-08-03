package udpp

import (
	"context"
	"fmt"
	"net"
)

var ErrPeerNotFound = fmt.Errorf("peer not found")

type Peer struct {
	ID       string
	Endpoint *net.UDPAddr
}

func NewPeer(id string, endpoint *net.UDPAddr) *Peer {
	return &Peer{ID: id, Endpoint: endpoint}
}

func FindPeer(ctx context.Context, id string) (*Peer, error) {
	endpoint := rdb.Get(ctx, id).Val()
	if endpoint == "" {
		return nil, ErrPeerNotFound
	}
	addr, err := net.ResolveUDPAddr("udp", endpoint)
	if err != nil {
		return nil, err
	}
	return NewPeer(id, addr), nil
}
