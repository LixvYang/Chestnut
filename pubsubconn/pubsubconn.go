// Package pubsubconn provides pubsubconn for chestnut.
package pubsubconn

import (
	chestnutpb "github.com/lixvyang/chestnut/pb"
)

type Chain interface {
	HandleTrx(trx *chestnutpb.Trx) error
	HandleBlock(block *chestnutpb.Block) error
}

type PubsubConn interface {
	JoinChannel(cId string, chain Chain) error
	Pubsub(data []byte) error
}