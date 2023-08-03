package udpp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Stream struct {
	net.Conn
	hostId  string
	peerId  string
	timeout int64 // Maximum idle time (in seconds)
	ts      int64 // Last received time
}

func NewStream(conn net.Conn, hostId, peerId string, timeout int64) *Stream {
	return &Stream{
		Conn:    conn,
		hostId:  hostId,
		peerId:  peerId,
		timeout: timeout,
		ts:      time.Now().Unix(),
	}
}

func (s *Stream) Read(p []byte) (n int, err error) {
	s.ts = time.Now().Unix()
	return s.Conn.Read(p)
}

func (s *Stream) Write(p []byte) (n int, err error) {
	return s.Conn.Write(p)
}

// Expired Checks if the connection is inactive.
func (s *Stream) Expired() bool {
	return time.Now().Unix()-s.ts > s.timeout
}

// Join provides a method for bridging two connections and facilitating data transmission
// between them. Note that both connections are closed after the function returns.
func (s *Stream) Join(other net.Conn) (err error) {
	var wg sync.WaitGroup
	wg.Add(2)

	pipe := func(to net.Conn, from net.Conn) {
		defer wg.Done()
		defer to.Close()
		defer from.Close()
		_, errCp := io.Copy(to, from)
		err = errors.Join(err, errCp)
	}

	go pipe(s, other)
	go pipe(other, s)
	wg.Wait()
	return
}

func (s *Stream) Ping(ctx context.Context) error {
	if err := rdb.Publish(ctx, s.peerId, s.hostId).Err(); err != nil {
		return err
	}
	buf := make([]byte, len(ping))
	// We set a short deadline here, so that
	// we can retry quickly after an incorrect port prediction.
	if err := s.Conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return err
	}
	defer s.Conn.SetReadDeadline(time.Time{})

	_, err := io.ReadFull(s.Conn, buf)
	if err != nil {
		return err
	}
	if !bytes.Equal(buf, ping) {
		return fmt.Errorf("invalid ping packet received")
	}
	return nil
}
