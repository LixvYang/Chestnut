// Package p2p provides p2p connectivity for chestnut.
package p2p

import (
	"context"
	"fmt"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	chestnutpb "github.com/lixvyang/chestnut/pb"
	"google.golang.org/protobuf/proto"
)

type PSPing struct {
	Topic *pubsub.Topic
	Subscription *pubsub.Subscription
	PeerId peer.ID
	ps *pubsub.PubSub
	ctx context.Context
}

type PingResult struct {
	Seqnum  int32
	Req_at  int64
	Resp_at int64
}

var ping_log = logging.Logger("ping")
var errCh chan error

func NewPSPingService(ctx context.Context, ps *pubsub.PubSub, peerid peer.ID) *PSPing {
	psping := &PSPing{PeerId: peerid, ps: ps, ctx: ctx}
	return psping
}

func (p *PSPing) EnablePing() error {
	peerid := p.PeerId.Pretty()
	var err error
	topicid := fmt.Sprint("PSPing: %s", peerid)
	p.Topic, err = p.ps.Join(topicid)
	if err != nil {
		ping_log.Infof("Enable PSPing channel <%s> failed", topicid)
		return err
	} else {
		ping_log.Infof("Enable PSPing channel <%s> done", topicid)
	}

	p.Subscription, err = p.Topic.Subscribe()
	if err != nil {
		ping_log.Fatalf("Subscribe PSPing channel <%s> failed", topicid)
		ping_log.Fatalf(err.Error())
		return err
	} else {
		ping_log.Infof("Subscribe PSPing channel <%s> done", topicid)
	}
	
	return nil
}

func (p *PSPing) handlePingRequest(pingresult *map[[32]byte]*PingResult) error {
	count := 0
	for {
		pingreqmsg, err := p.Subscription.Next(p.ctx)
		if err == nil {
			if pingreqmsg.ReceivedFrom != p.PeerId {
				var pspingresp chestnutpb.PSPing
				if err := proto.Unmarshal(pingreqmsg.Data, &pspingresp); err != nil {
					return err
				}
				if pspingresp.IsResp {
					var payload [32]byte
					copy(payload[:], pspingresp.Payload[0:32])
					_, ok := (*pingresult)[payload]
					if ok {
						(*pingresult)[payload].Resp_at = time.Now().UnixNano()
						count++
						if count == 10 {
							errCh <- nil
							return nil
						}
					}
				}
			}
		} else {
			ping_log.Error(err.Error())
			return err
		}
	}
}