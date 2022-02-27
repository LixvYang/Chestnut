// Package pubsubconn provides pubsubconn for chestnut.
package pubsubconn

import (
	"context"

	logging "github.com/ipfs/go-log/v2"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	chestnutpb "github.com/lixvyang/chestnut/pb"
	"google.golang.org/protobuf/proto"
)

var channel_log = logging.Logger("chan")

type P2pPubSubConn struct {
	Cid string
	Topic *pubsub.Topic
	Subscription *pubsub.Subscription
	chain Chain
	ps *pubsub.PubSub
	nodename string
	Ctx context.Context
}

func InitP2pPubSubConn(ctx context.Context, ps *pubsub.PubSub, nodename string) *P2pPubSubConn {
	return &P2pPubSubConn{Ctx: ctx, ps: ps, nodename: nodename}
}

func (psconn *P2pPubSubConn) JoinChannel(cId string, chain Chain) error {
	psconn.Cid = cId
	psconn.chain = chain

	var err error

	//TODO:Share the ps
	psconn.Topic, err = psconn.ps.Join(cId)
	if err != nil {
		channel_log.Infof("Join <%s> failed", cId)
		return err
	} else {
		channel_log.Infof("Join <%s> success", cId)
	}

	psconn.Subscription, err = psconn.Topic.Subscribe()
	if err != nil {
		channel_log.Fatalf("Subscribe <%s> failed", cId)
		channel_log.Fatalf(err.Error())
		return err
	} else {
		channel_log.Infof("Subscribe <%s> done", cId)
	}
	go psconn.handleGroupChannel()
	return nil
}

func (psconn *P2pPubSubConn) Publish(data []byte) error {
	return psconn.Topic.Publish(psconn.Ctx, data)
}

func (psconn *P2pPubSubConn) handleGroupChannel() error {
	for {
		msg, err := psconn.Subscription.Next(psconn.Ctx)
		if err == nil {
			var pkg chestnutpb.Package
			err = proto.Unmarshal(msg.Data, &pkg)
			if err == nil {
				if pkg.Type == chestnutpb.PackageType_BLOCK {
					// is block
					var blk *chestnutpb.Block
					blk = &chestnutpb.Block{}
					err := proto.Unmarshal(pkg.Data, blk)
					if err == nil {
						psconn.chain.HandleBlock(blk)
					} else {
						channel_log.Warning(err.Error())
					}
				} else if pkg.Type == chestnutpb.PackageType_TRX{
					var trx *chestnutpb.Trx
					trx = &chestnutpb.Trx{}
					err := proto.Unmarshal(pkg.Data, trx)
					if err == nil {
						psconn.chain.HandleTrx(trx)
					} else {
						channel_log.Warning(err.Error())
					}
				}
			} else {
				channel_log.Warningf(err.Error())
			}
		} else {
			channel_log.Warningf(err.Error())
			return err
		}
	}
}