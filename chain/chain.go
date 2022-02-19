// Package chain provides chain for chestnut.
package chain

import (
	"sync"

	chestnutpb "github.com/lixvyang/chestnut/pb"
)

type Chain struct {
	nodename          string
	group             *Group
	userChannelId     string
	producerChannelId string
	trxMgrs           map[string]string
	ProducerPool      map[string]*chestnutpb.ProducerItem

	Syncer    *Syncer
	Consensus Consensus
	statusmu  sync.RWMutex
	groupId   string
}
