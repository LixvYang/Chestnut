// Package chain provides chain for chestnut.
package chain

import (
	chestnutpb "github.com/lixvyang/chestnut/pb"
)

type ChainMolassesIface interface {
	GetUserTrxMgr() *TrxMgr
	GerProducerTrxMgr() *TrxMgr
	UpdChainInfo(height int64, blockId string) error
	UpdProducerList()
	CreateConsensus()
	IsSyncerReady() bool
	SyncBackward(block *chestnutpb.Block) error
}