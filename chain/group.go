// Package chain provides chain for chestnut.
package chain

import (
	chestnutpb "github.com/lixvyang/chestnut/pb"
	logging "github.com/ipfs/go-log/v2"
)

const (
	USER_CHANNEL_PREFIX = "user_channel_"
	PRODUCER_CHANNEL_PREFIX = "prod_channel_"
)

var group_log = logging.Logger("group")

type Group struct {
	// Group Item
	Item *chestnutpb.GroupItem
	ChainCtx *Chain
}

func (grp *Group) Init(item *chestnutpb.GroupItem)  {
	groupMgr_log.Debugf("<%s> Init called", item.GroupId)
	grp.ChainCtx = &Chain{}
	// grp.ChainCtx.Init(item)

}