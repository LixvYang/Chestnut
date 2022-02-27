// Package pubsubconn provides pubsubconn for chestnut.
package pubsubconn

import (
	chestnutpb "github.com/lixvyang/chestnut/pb"
)

type Chain interface {
	HandleTrx(trx *chestnutpb.Trx) error
	HandleBlock(block *chestnutpb.Block) error
}

type PubSubConn interface {
	JoinChannel(cId string, chain Chain) error
	Publish(data []byte) error
}