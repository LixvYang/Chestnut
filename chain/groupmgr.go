// Package chain provides chain for chestnut.
package chain

import (
	logging "github.com/ipfs/go-log/v2"
	"github.com/lixvyang/chestnut/storage"
)
type GroupMgr struct {
	dbMgr *storage.DbMgr
	Groups map[string]*Group
}

var groupmgr *GroupMgr

var groupMgr_log = logging.Logger("groupmgr")

func GetGroupMgr() *GroupMgr {
	return groupmgr 
}