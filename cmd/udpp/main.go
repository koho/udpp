package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/koho/udpp"
	"github.com/koho/udpp/config"
)

var cfgFile string

func init() {
	rootCmd.Flags().StringVarP(&cfgFile, "config", "c", "./config.yml", "config file path")
}

var rootCmd = &cobra.Command{
	Use:   "udpp",
	Short: "A Point-to-Point UDP Tunnel.",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		var cfg = config.Default()

		if err = cfg.Load(cfgFile); err != nil {
			return err
		}
		if cfg.Local == "" {
			return fmt.Errorf("local address is not specified")
		}
		localAddr, err := net.ResolveUDPAddr("udp", cfg.Local)
		if err != nil {
			return err
		}
		var bindAddr *net.UDPAddr
		if cfg.Peer.Bind != "" {
			bindAddr, err = net.ResolveUDPAddr("udp", cfg.Peer.Bind)
			if err != nil {
				return err
			}
		}

		if err = udpp.Setup(cfg.Server); err != nil {
			return err
		}
		fmt.Printf("Node ID: %s\n", cfg.ID)

		if cfg.Peer.ID == "" {
			fmt.Printf("Connection: %s -> %s\n", "any", localAddr)
		} else {
			fmt.Printf("Peer ID: %s\n", cfg.Peer.ID)
			fmt.Printf("Connection: %s -> %s\n", bindAddr, localAddr)
		}

		var lastErr error
		for {
			if cfg.Peer.ID != "" {
				err = access(&cfg, localAddr, bindAddr)
			} else {
				err = serve(&cfg, localAddr)
			}
			if errors.Is(err, udpp.ErrPeerNotFound) {
				if !errors.Is(lastErr, udpp.ErrPeerNotFound) {
					log.Printf("waiting for peer %s\n", cfg.Peer.ID)
				}
			} else if !errors.Is(err, udpp.ErrNodeInactive) {
				log.Println(err)
			}
			lastErr = err
			time.Sleep(5 * time.Second)
		}
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
