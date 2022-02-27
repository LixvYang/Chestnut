// Package chain provides chain for chestnut.
package chain

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	logging "github.com/ipfs/go-log/v2"
	chestnutpb "github.com/lixvyang/chestnut/pb"
	"github.com/lixvyang/chestnut/storage"
)
type GroupMgr struct {
	dbMgr *storage.DbMgr
	Groups map[string]*Group
}

var groupMgr *GroupMgr
var groupMgr_log = logging.Logger("groupmgr")

func GetGroupMgr() *GroupMgr {
	return groupMgr 
}

// TODO:singlaton
func InitGroupMgr(dbMgr *storage.DbMgr) *GroupMgr {
	groupMgr_log.Debug("InitGroupMgr called")
	groupMgr = &GroupMgr{dbMgr: dbMgr}
	groupMgr.Groups = make(map[string]*Group)
	return groupMgr
}


// load and group add start syncing
func (groupmgr *GroupMgr) SyncAllGroup() error {
	groupMgr_log.Debug("SyncAllGroup called")

	// open all groups
	groupItemsBytes, err := groupmgr.dbMgr.GetGroupsBytes()
	if err != nil {
		return err
	}

	for _, b := range groupItemsBytes {
		var group *Group
		group = &Group{}

		var item *chestnutpb.GroupItem
		item = &chestnutpb.GroupItem{}

		proto.Unmarshal(b, item)
		group.Init(item)
		if err == nil {
			groupMgr_log.Debugf("Start sync group: %s", item.GroupId)
			go group.StopSync()
			groupmgr.Groups[item.GroupId] = group
		} else {
			groupMgr_log.Fatalf("can't sync group: %s", item.GroupId)
			groupMgr_log.Fatalf(err.Error())
		}
	}
	return nil
}

func (groupmgr *GroupMgr) StopSyncAllGroup() error {
	groupMgr_log.Debug("StopSyncAllGroup called")
	return nil
}

func (groupmgr *GroupMgr) Release()  {
	groupMgr_log.Debug("Release called")
	for groupId, group := range groupMgr.Groups {
		groupMgr_log.Debugf("group: <%s> teardown", groupId)
		group.TearDown()
	}
	// close ctx db
	groupmgr.dbMgr.CloseDb()
}

func (groupmgr *GroupMgr) GetGroupItem(groupId string) (*chestnutpb.GroupItem, error) {
	if grp, ok := groupmgr.Groups[groupId]; ok {
		return grp.Item, nil
	}
	return nil, fmt.Errorf("group not exist: %s", groupId)
}