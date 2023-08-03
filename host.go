package udpp

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/fatedier/frp/pkg/nathole"

	"github.com/koho/udpp/config"
)

var StreamExists = fmt.Errorf("stream already exists")

type Host struct {
	ID        string
	Endpoint  *net.UDPAddr        // Public network endpoint
	LocalAddr *net.UDPAddr        // Local network endpoint
	NAT       *nathole.NatFeature // NAT information of the host

	mu      sync.Mutex
	streams map[string]*Stream
	timeout int64 // Maximum idle time of a stream (in seconds)
	ctx     context.Context
	cancel  context.CancelFunc
	done    chan struct{} // Channel that will be closed when the host has cleanly shutdown
}

// NewHost maps a local address to a public address using STUN server.
func NewHost(ctx context.Context, id string, opts ...config.Option) (*Host, error) {
	var cfg = config.Default()
	if err := cfg.Apply(opts...); err != nil {
		return nil, err
	}

	// Discover NAT network information
	nat, err := Discover(cfg.Stun)
	if err != nil {
		return nil, err
	}
	endpoint := nat.RemoteAddrs[len(nat.RemoteAddrs)-1]

	if nat.Feature.Behavior == nathole.BehaviorPortChanged {
		// This only works with port auto-increment
		endpoint.Port++
	}
	if err = rdb.Set(ctx, id, endpoint.String(), time.Duration(cfg.Timeout)*time.Second).Err(); err != nil {
		return nil, err
	}
	host := &Host{
		ID:        id,
		Endpoint:  endpoint,
		LocalAddr: nat.LocalAddr,
		NAT:       nat.Feature,
		timeout:   cfg.Timeout,
		streams:   make(map[string]*Stream),
		done:      make(chan struct{}),
	}
	host.ctx, host.cancel = context.WithCancel(ctx)
	// Start a worker to maintain the host state
	go host.keepalive()
	return host, nil
}

// NewStream creates an udp connection to the given peer.
func (h *Host) NewStream(peer *Peer) (*Stream, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.ctx.Err() != nil {
		return nil, h.ctx.Err()
	}
	if old, ok := h.streams[peer.ID]; ok {
		return old, StreamExists
	}
	conn, err := net.DialUDP("udp", h.LocalAddr, peer.Endpoint)
	if err != nil {
		return nil, err
	}
	stream := NewStream(conn, h.ID, peer.ID, h.timeout)
	// The ping packet is required to create a peer entity in firewall.
	if _, err = stream.Write(ping); err != nil {
		return nil, err
	}
	h.streams[peer.ID] = stream
	return stream, nil
}

func (h *Host) mustCreateStream(peer *Peer) (*Stream, error) {
	for {
		stream, err := h.NewStream(peer)
		if err != nil {
			// Replace old stream when the peer restarted.
			if errors.Is(err, StreamExists) {
				h.mu.Lock()
				log.Printf("[%s] replace stream with %s\n", peer.ID, peer.Endpoint)
				stream.Close()
				delete(h.streams, peer.ID)
				h.mu.Unlock()
				time.Sleep(2 * time.Second)
				continue
			}
			return nil, err
		}
		return stream, nil
	}
}

// Serve listens for new peer connection.
func (h *Host) Serve(addr *net.UDPAddr) error {
	sub := rdb.Subscribe(h.ctx, h.ID)
	defer sub.Close()
	for {
		select {
		case req := <-sub.Channel():
			peer, err := FindPeer(h.ctx, req.Payload)
			if err != nil {
				log.Println(err)
				continue
			}
			stream, err := h.mustCreateStream(peer)
			if err != nil {
				log.Println(err)
				continue
			}
			go func() {
				conn, err := net.DialUDP("udp", nil, addr)
				if err != nil {
					log.Println(err)
					return
				}
				defer conn.Close()
				log.Printf("[%s] new connection (%s, %s) -> (%s, %s)\n",
					peer.ID, conn.LocalAddr(), conn.RemoteAddr(), stream.LocalAddr(), stream.RemoteAddr())
				if err = stream.Join(conn); err != nil {
					log.Println(err)
				}
				log.Printf("[%s] stream closed\n", peer.ID)
			}()
		case <-h.ctx.Done():
			return h.ctx.Err()
		}
	}
}

func (h *Host) keepalive() {
	defer close(h.done)
	timer := time.NewTimer(60 * time.Second)
	for {
		select {
		case <-h.Expired():
			return
		case <-timer.C:
			alive := false
			h.mu.Lock()
			for id, s := range h.streams {
				if s.Expired() {
					log.Printf("[%s] close stream due to inactivity\n", id)
					s.Close()
					delete(h.streams, id)
				} else {
					alive = true
				}
			}
			h.mu.Unlock()
			if alive {
				if err := rdb.Set(h.ctx, h.ID, h.Endpoint.String(), time.Duration(h.timeout)*time.Second).Err(); err != nil {
					log.Println(err)
				}
				timer.Reset(60 * time.Second)
			} else {
				h.cancel()
				return
			}
		}
	}
}

func (h *Host) Expired() <-chan struct{} {
	return h.ctx.Done()
}

func (h *Host) Close() (err error) {
	h.mu.Lock()
	for id, s := range h.streams {
		err = errors.Join(err, s.Close())
		delete(h.streams, id)
	}
	h.mu.Unlock()
	h.cancel()
	<-h.done
	return
}
