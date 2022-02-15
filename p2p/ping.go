// Package p2p provides p2p connectivity for chestnut.
package p2p

import (
	"bytes"
	"context"
	"errors"
	"io"
	"time"

	u "github.com/ipfs/go-ipfs-util"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
)

var pinglog = logging.Logger("ping")

const (
	PingSize = 32
	PingID = "/chestnut/ping/1.0.0"
	pingTimeout = time.Second * 60
)

type PingService struct {
	Host host.Host
}

func NewPingService(h host.Host) *PingService {
	ps := &PingService{Host: h}
	h.SetStreamHandler(PingID, ps.PingHandler)
	return ps	
}

func (ps *PingService) PingHandler(s network.Stream)  {
	buf := make([]byte, PingSize)

	errCh := make(chan error, 1)
	defer close(errCh)
	timer := time.NewTimer(pingTimeout)
	defer timer.Stop()

	go func() {
		select {
		case <-timer.C:
			pinglog.Debug("ping timeout")
		case err, ok := <- errCh:
			if ok {
				pinglog.Debug(err)
			} else {
				pinglog.Error("ping loop failed without error")
			}
		}
		s.Reset()
	}()

	for {
		_, err := io.ReadFull(s, buf)
		if err != nil {
			errCh <- err
			return 
		}
		_, err = s.Write(buf)
		if err != nil {
			errCh <- err
			return 
		}
		timer.Reset(pingTimeout)
	}
}

type Result struct {
	RTT time.Duration
	Error error
}

func (ps *PingService) Ping(ctx context.Context, p peer.ID) <-chan Result {
	return Ping(ctx, ps.Host, p)
}

func Ping(ctx context.Context, h host.Host, p peer.ID) <-chan Result {
	s, err := h.NewStream(ctx, p, PingID)
	if err != nil {
		ch := make(chan Result, 1)
		ch <- Result{Error: err}
		close(ch)
		return ch
	}

	ctx, cancel := context.WithCancel(ctx)

	out := make(chan Result)
	go func() {
		defer close(out)
		defer cancel()

		for ctx.Err() == nil {
			var res Result
			res.RTT, res.Error = ping(s)

			// canceled, ignore everything.
			if ctx.Err() != nil {
				return 
			}

			// No error, record the RTT
			if res.Error == nil {
				h.Peerstore().RecordLatency(p, res.RTT)
			}

			select {
			case out <- res:
			case <- ctx.Done():
				return 
			}
		}
	}()

	go func() {
		// forces the ping to abort
		<- ctx.Done()
		s.Reset()
	}()
	
	return out
}

func ping(s network.Stream) (time.Duration, error) {
	buf := make([]byte, PingSize)
	u.NewTimeSeededRand().Read(buf)

	before := time.Now()
	_, err := s.Write(buf)
	if err != nil {
		return 0, err
	}

	rbuf := make([]byte, PingSize)
	_, err = io.ReadFull(s, rbuf)
	if err != nil {
		return 0, err
	}

	if !bytes.Equal(buf, rbuf) {
		return 0, errors.New("ping packet was incorrect")
	}
	return time.Since(before), nil
}