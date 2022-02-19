// Package chain provides chain for chestnut.
package chain

import (
	chestnutpb "github.com/lixvyang/chestnut/pb"
)

type Producer interface {
	Init(item *chestnutpb.GroupItem, nodename string, iface ChainMolassesIface)
	AddTrx(trx *chestnutpb.Trx)
	AddBlockToPool(block *chestnutpb.Block)
	GetBlockForward(trx *chestnutpb.Trx) error
	GetBlockBackward(trx *chestnutpb.Trx) error
	AddProducedBlock(trx *chestnutpb.Trx) error
	AddBlock(block *chestnutpb.Block) error
}