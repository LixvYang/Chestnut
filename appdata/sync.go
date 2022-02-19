// Package appdata provides storage for chestnut.
package appdata

import (
	logging "github.com/ipfs/go-log/v2"
	"github.com/lixvyang/chestnut/storage"
	"github.com/lixvyang/chestnut/chain"
	chestnutpb "github.com/lixvyang/chestnut/pb"
)

var appsynclog = logging.Logger("appsync")

type AppSync struct {
	appdb *AppDb
	dbmgr *storage.DbMgr
	groupmgr *chain.GroupMgr
	apiroot string
	nodename string
}

func NewAppSyncAgent(apiroot string, nodename string, appdb *AppDb, dbmgr *storage.DbMgr) *AppSync {
	groupmgr := chain.GetGroupMgr()
	appsync := &AppSync{appdb, dbmgr, groupmgr, apiroot, nodename}
	return appsync
}

func (appsync *AppSync) GetGroups() []*chestnutpb.GroupItem {
	var items []*chestnutpb.GroupItem
	for _, grp := range appsync.groupmgr.Groups {
		items = append(items, grp.Item)
	}
	return items
}

// func (appsync *AppSync) ParseBlockTrx(groupid string, block *chestnutpb.Block) ([]*chestnutpb.Block, error) {
// 	appsynclog.Infof("ParseBlockTrxs %d trx(s) on group %s", len(block.Trxs), groupid)
// 	if err := appsync.appdb.AddMetaByTrx(block.BlockId, groupid, block.Trxs); err != nil {
// 		appsynclog.Errorf("ParseBlockTrxs on group %s err: ", groupid, err)
// 	}
// 	return appsync.dbmgr.GetSubBlock(block.BlockId, appsync.nodename)
// }

