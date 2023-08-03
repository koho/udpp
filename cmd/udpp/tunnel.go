package main

import (
	"context"
	"log"
	"net"

	"github.com/koho/udpp"
	"github.com/koho/udpp/config"
)

func serve(cfg *config.Config, localAddr *net.UDPAddr) error {
	host, err := udpp.NewHost(context.Background(), cfg.ID, config.Timeout(cfg.Timeout), config.Stun(cfg.Stun))
	if err != nil {
		return err
	}
	defer host.Close()
	return host.Serve(localAddr)
}

func access(cfg *config.Config, localAddr, bindAddr *net.UDPAddr) error {
	peer, err := udpp.FindPeer(context.Background(), cfg.Peer.ID)
	if err != nil {
		return err
	}

	log.Printf("found peer %s (%s)\n", cfg.Peer.ID, peer.Endpoint)
	host, err := udpp.NewHost(context.Background(), cfg.ID, config.Timeout(cfg.Timeout), config.Stun(cfg.Stun))
	if err != nil {
		return err
	}
	defer host.Close()
	log.Printf("created node %s (%s)\n", cfg.ID, host.Endpoint)

	conn, err := net.DialUDP("udp", bindAddr, localAddr)
	if err != nil {
		return err
	}
	defer conn.Close()
	stream, err := host.NewStream(peer)
	if err != nil {
		return err
	}
	defer stream.Close()

	log.Printf("connecting to peer %s (%s)\n", peer.ID, peer.Endpoint)
	if err = stream.Ping(context.Background()); err != nil {
		return err
	}
	log.Printf("[%s] connection established (%s, %s) -> (%s, %s)\n",
		peer.ID, bindAddr, localAddr, stream.LocalAddr(), stream.RemoteAddr())
	defer log.Printf("[%s] connection closed\n", peer.ID)
	return stream.Join(conn)
}
