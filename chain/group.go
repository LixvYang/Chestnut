// Package chain provides chain for chestnut.
package chain

import (
	"bytes"
	"encoding/hex"
	"errors"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/lixvyang/chestnut/nodectx"
	chestnutpb "github.com/lixvyang/chestnut/pb"
	"google.golang.org/protobuf/proto"
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
	grp.ChainCtx.Init(grp)

	grp.ChainCtx.UpdProducerList()
	grp.ChainCtx.CreateConsensus()
	group_log.Infof("Group <%s> initialed", grp.Item.GroupId)
}

// teardown group
func (grp *Group) TearDown() {
	groupMgr_log.Debugf("<%s> TearDown called", grp.Item.GroupId)
	if grp.ChainCtx.Syncer.Status == SYNCING_BACKWARD || grp.ChainCtx.Syncer.Status == SYNCING_FORWARD {
		grp.ChainCtx.Syncer.stopWaitBlock()
	}

	group_log.Infof("Group <%s> teardown", grp.Item.GroupId)
}

func (grp *Group) CreateGrp(item *chestnutpb.GroupItem) error {
	group_log.Debugf("<%s> GreateGrp called", item.GroupId)
	grp.Init(item)

	err := nodectx.GetDbMgr().AddGensisBlock(item.GenesisBlock, grp.ChainCtx.nodename)
	if err != nil {
		return err
	}

	group_log.Debugf("<%s> add owner as the first producer", grp.Item.GroupId)
	//add owner as the first producer
	var pItem *chestnutpb.ProducerItem
	pItem = &chestnutpb.ProducerItem{}
	pItem.GroupId = item.GroupId
	pItem.GroupOwnerPubkey = item.OwnerPubKey
	pItem.ProducerPubkey = item.OwnerPubKey

	var buffer bytes.Buffer
	buffer.Write([]byte(pItem.GroupId))
	buffer.Write([]byte(pItem.ProducerPubkey))
	buffer.Write([]byte(pItem.GroupOwnerPubkey))
	hash := Hash(buffer.Bytes())

	ks := nodectx.GetNodeCtx().Keystore
	signature, err := ks.SignByKeyName(item.GroupId, hash)
	if err != nil {
		return err
	}

	pItem.GroupOwnerSign = hex.EncodeToString(signature)
	pItem.Memo = "Owner Registate as the first oroducer"
	pItem.TimeStamp = time.Now().UnixNano()

	err = nodectx.GetDbMgr().AddProducer(pItem, grp.ChainCtx.nodename)
	if err != nil {
		return err
	}

	// reload producers
	grp.ChainCtx.UpdProducerList()
	grp.ChainCtx.CreateConsensus()

	group_log.Infof("Group <%s> created", grp.Item.GroupId)
	return nodectx.GetDbMgr().AddGroup(grp.Item)
}


func (grp *Group) DelGrp() error {
	group_log.Debugf("<%s> DelGrp called", grp.Item.GroupId)
	if grp.Item.UserSignPubkey != grp.Item.OwnerPubKey {
		err := errors.New("You can not 'delete' group created by others, use 'leave' instead")
		return err
	}

	err := grp.clearGroup()
	if err != nil {
		return err
	}

	group_log.Infof("Group <%s> deleted", grp.Item.GroupId)
	return nodectx.GetDbMgr().RmGroup(grp.Item)
}

func (grp *Group) LeaveGrp() error {
	group_log.Debugf("<%s> LeaveGrp called", grp.Item.GroupId)
	if grp.Item.UserSignPubkey == grp.Item.OwnerPubKey {
		err := errors.New("Group creator can not leave the group they created, use 'delete' instead")
		return err
	}

	err := grp.clearGroup()
	if err != nil {
		return err
	}

	group_log.Infof("Group <%s> leaved", grp.Item.GroupId)

	return nodectx.GetDbMgr().RmGroup(grp.Item)
}


func (grp *Group) clearGroup() error {

	//remove all group blocks (both cached and normal)

	//remove all group producers

	//remove all group trx

	//remove all group POST

	//remove all group CONTENT

	//remove all group Auth

	//remove all group Announce

	//remove all group schema

	return nil
}

func (grp *Group) StopSync() error {
	group_log.Debugf("<%s> StopSync called", grp.Item.GroupId)
	if grp.ChainCtx.Syncer.Status == SYNCING_BACKWARD || grp.ChainCtx.Syncer.Status == SYNCING_FORWARD {
		grp.ChainCtx.StopSync()
	}

	group_log.Infof("Group <%s> stop sync", grp.Item.GroupId)
	return nil
}


func (grp *Group) GetGroupCtn(filter string) ([]*chestnutpb.PostItem, error) {
	group_log.Debugf("<%s> GetGroupCtn called", grp.Item.GroupId)
	return nodectx.GetDbMgr().GetGrpCtnt(grp.Item.GroupId, filter, grp.ChainCtx.nodename)
}

func (grp *Group) GetBlock(blockId string) (*chestnutpb.Block, error) {
	group_log.Debugf("<%s> GetBlock called", grp.Item.GroupId)
	return nodectx.GetDbMgr().GetBlock(blockId, false, grp.ChainCtx.nodename)
}

func (grp *Group) GetTrx(trxId string) (*chestnutpb.Trx, error) {
	group_log.Debugf("<%s> GetTrx called", grp.Item.GroupId)
	return nodectx.GetDbMgr().GetTrx(trxId, grp.ChainCtx.nodename)
}

func (grp *Group) GetBlockedUser() ([]*chestnutpb.DenyUserItem, error) {
	group_log.Debugf("<%s> GetBlockedUser called", grp.Item.GroupId)
	return nodectx.GetDbMgr().GetBlkedUsers(grp.ChainCtx.nodename)
}

func (grp *Group) GetProducers() ([]*chestnutpb.ProducerItem, error) {
	group_log.Debugf("<%s> GetProducers called", grp.Item.GroupId)
	return nodectx.GetDbMgr().GetProducers(grp.Item.GroupId, grp.ChainCtx.nodename)
}

func (grp *Group) GetAnnouncedUser() ([]*chestnutpb.AnnounceItem, error) {
	group_log.Debugf("<%s> GetAnnouncedUser called", grp.Item.GroupId)
	return nodectx.GetDbMgr().GetAnnouncedUsersByGroup(grp.Item.GroupId, grp.ChainCtx.nodename)
}

func (grp *Group) GetAnnouncedProducer() ([]*chestnutpb.AnnounceItem, error) {
	group_log.Debugf("<%s> GetAnnouncedProducer called", grp.Item.GroupId)
	return nodectx.GetDbMgr().GetAnnounceProducersByGroup(grp.Item.GroupId, grp.ChainCtx.nodename)
}

func (grp *Group) UpdAnnounce(item *chestnutpb.AnnounceItem) (string, error) {
	group_log.Debugf("<%s> UpdAnnounce called", grp.Item.GroupId)
	return grp.ChainCtx.Consensus.User().UpdAnnounce(item)
}

func (grp *Group) UpdBlkList(item *chestnutpb.DenyUserItem) (string, error) {
	group_log.Debugf("<%s> UpdBlkList called", grp.Item.GroupId)
	return grp.ChainCtx.Consensus.User().UpdBlkList(item)
}

func (grp *Group) PostToGroup(content proto.Message) (string, error) {
	group_log.Debugf("<%s> PostToGroup called", grp.Item.GroupId)
	return grp.ChainCtx.Consensus.User().PostToGroup(content)
}

func (grp *Group) UpdProducer(item *chestnutpb.ProducerItem) (string, error) {
	group_log.Debugf("<%s> UpdProducer called", grp.Item.GroupId)
	return grp.ChainCtx.Consensus.User().UpdProducer(item)
}

func (grp *Group) UpdSchema(item *chestnutpb.SchemaItem) (string, error) {
	group_log.Debugf("<%s> UpdSchema called", grp.Item.GroupId)
	return grp.ChainCtx.Consensus.User().UpdSchema(item)
}

func (grp *Group) IsProducerAnnounced(producerSignPubkey string) (bool, error) {
	group_log.Debugf("<%s> IsProducerAnnounced called", grp.Item.GroupId)
	return nodectx.GetDbMgr().IsProducerAnnounced(grp.Item.GroupId, producerSignPubkey, grp.ChainCtx.nodename)
}