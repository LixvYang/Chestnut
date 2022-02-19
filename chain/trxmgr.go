// Package chain provides chain for chestnut.
package chain

import (
	chestnutpb "github.com/lixvyang/chestnut/pb"
	"github.com/lixvyang/chestnut/pubsubconn"
)

type TrxMgr struct {
	nodename string
	groupItem chestnutpb.GroupItem
	psconn pubsubconn.PubsubConn
	groupId string
}