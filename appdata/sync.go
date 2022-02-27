// Package appdata provides storage for chestnut.
package appdata

import (
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/lixvyang/chestnut/chain"
	chestnutpb "github.com/lixvyang/chestnut/pb"
	"github.com/lixvyang/chestnut/storage"
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

func (appsync *AppSync) ParseBlockTrx(groupid string, block *chestnutpb.Block) ([]*chestnutpb.Block, error) {
	appsynclog.Infof("ParseBlockTrxs %d trx(s) on group %s", len(block.Trxs), groupid)
	if err := appsync.appdb.AddMetaByTrx(block.BlockId, groupid, block.Trxs); err != nil {
		appsynclog.Errorf("ParseBlockTrxs on group %s err: ", groupid, err)
	}
	return appsync.dbmgr.GetSubBlock(block.BlockId, appsync.nodename)
}

func (appsync *AppSync) RunSync(groupid string, lastBlockedId string, newBlockId string)  {
	var blocks []*chestnutpb.Block
	subblocks, err := appsync.dbmgr.GetSubBlock(lastBlockedId, appsync.nodename)
	if err == nil {
		blocks = append(blocks, subblocks...)
		for {
			if len(blocks) == 0 {
				appsynclog.Infof("no new blocks, skip sync")
				break
			}
			var blk *chestnutpb.Block
			blk, blocks = blocks[0], blocks[1:]
			newsubblocks, err := appsync.ParseBlockTrx(groupid, blk)
			if err == nil {
				blocks = append(blocks, newsubblocks...)
			} else {
				appsynclog.Errorf("ParseBlockTrxs error %s", err)
			}
		}
	} else {
		appsynclog.Errorf("db read err: %s", err)
	}
}

func (appsync *AppSync) Start(interval int)  {
	go func() {
		for {
			groups := appsync.GetGroups()
			for _, groupitem := range groups {
				lastBlockId, err := appsync.appdb.GetGroupStatus(groupitem.GroupId, "HighestBlockId")
				if err == nil {
					if lastBlockId == "" {
						lastBlockId = groupitem.GenesisBlock.BlockId
					}
					if lastBlockId != groupitem.HighestBlockId {
						appsync.RunSync(groupitem.GroupId, lastBlockId, groupitem.HighestBlockId)
					}
				} else {
					appsynclog.Errorf("sync group: %s Get HeightBlockId err %s", groupitem.GroupId, err)
				}
			}
			time.Sleep(time.Duration(interval) * time.Second)
		}
	}()
}

