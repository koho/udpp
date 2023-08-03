package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/koho/udpp"
	"github.com/koho/udpp/config"
)

func main() {
	var err error
	var cfg = config.Default()

	path := "config.yml"
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	if err = cfg.Load(path); err != nil {
		panic(err)
	}
	if cfg.Local == "" {
		panic(fmt.Errorf("local address is not specified"))
	}
	localAddr, err := net.ResolveUDPAddr("udp", cfg.Local)
	if err != nil {
		panic(err)
	}
	var bindAddr *net.UDPAddr
	if cfg.Peer.Bind != "" {
		bindAddr, err = net.ResolveUDPAddr("udp", cfg.Peer.Bind)
		if err != nil {
			panic(err)
		}
	}

	if err = udpp.Setup(cfg.Server); err != nil {
		panic(err)
	}
	fmt.Printf("Node ID: %s\n", cfg.ID)

	if cfg.Peer.ID == "" {
		fmt.Println("Mode: forward")
		fmt.Printf("Target: %s\n", localAddr)
	} else {
		fmt.Println("Mode: access")
		fmt.Printf("Peer: %s\n", cfg.Peer.ID)
		fmt.Printf("Listen: %s\n", bindAddr)
		fmt.Printf("From: %s\n", localAddr)
	}

	var lastErr error
	for {
		if cfg.Peer.ID != "" {
			err = access(&cfg, localAddr, bindAddr)
		} else {
			err = serve(&cfg, localAddr)
		}
		if errors.Is(err, udpp.PeerNotFound) {
			if !errors.Is(lastErr, udpp.PeerNotFound) {
				log.Printf("waiting for peer %s\n", cfg.Peer.ID)
			}
		} else {
			log.Println(err)
		}
		lastErr = err
		time.Sleep(5 * time.Second)
	}
}
