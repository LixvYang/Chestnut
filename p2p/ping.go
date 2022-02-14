// Package p2p provides p2p connectivity for chestnut.
package p2p

import (
	"io"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
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



