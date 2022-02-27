// Package chain provides chain for chestnut.
package chain

import (
	"encoding/hex"
	"errors"
	localcrypto "github.com/lixvyang/chestnut/crypto"
	logging "github.com/ipfs/go-log/v2"
	"github.com/lixvyang/chestnut/nodectx"
	chestnutpb "github.com/lixvyang/chestnut/pb"
	"google.golang.org/protobuf/proto"
)

var molauser_log = logging.Logger("user")

type MolassesUser struct {
	grpItem *chestnutpb.GroupItem
	nodename string
	cIface ChainMolassesIface
	groupId string
}

func (user *MolassesUser) Init(item *chestnutpb.GroupItem, nodename string, cIface ChainMolassesIface)  {
	molaproducer_log.Debugf("Init called")
	user.grpItem = item
	user.nodename = nodename
	user.cIface = cIface
	user.groupId = item.GroupId
	molaproducer_log.Infof("<%s> User created", user.groupId)
}

func (user *MolassesUser) UpdAnnounce(item *chestnutpb.AnnounceItem) (string, error) {
	molauser_log.Debugf("<%s> UpdAnnounce called called", user.groupId)
	return user.cIface.GetProducerTrxMgr().SendAnnounceTrx(item)
}

func (user *MolassesUser) UpdBlkList(item *chestnutpb.DenyUserItem) (string, error) {
	molauser_log.Debugf("<%s> UpdBlkList called", user.groupId)
	return user.cIface.GetProducerTrxMgr().SendUpdAuthTrx(item)
}

func (user *MolassesUser) UpdSchema(item *chestnutpb.SchemaItem) (string, error) {
	molauser_log.Debugf("<%s> UpdSchema called", user.groupId)
	return user.cIface.GetProducerTrxMgr().SendUpdSchemaTrx(item)
}
func (user *MolassesUser) UpdProducer(item *chestnutpb.ProducerItem) (string, error) {
	molauser_log.Debugf("<%s> UpdSchema called", user.groupId)
	return user.cIface.GetProducerTrxMgr().SendRegProducerTrx(item)
}

func (user *MolassesUser) PostToGroup(content proto.Message) (string, error) {
	molauser_log.Debugf("<%s> PostToGroup called", user.groupId)
	if user.cIface.IsSyncerReady() {
		return "", errors.New("can not post to group, group is in sycing or sync failed")
	}
	return user.cIface.GetProducerTrxMgr().PostAny(content)
}




func (user *MolassesUser) AddBlock(block *chestnutpb.Block) error {
	molauser_log.Debugf("<%s> AddBlock called", user.groupId)

	//check if block is already in chain
	isSaved, err := nodectx.GetDbMgr().IsBlockExist(block.BlockId, false, user.nodename)
	if err != nil {
		return err
	}

	if isSaved {
		return errors.New("Block already saved, ignore")
	}

	//check if block is in cache
	isCached, err := nodectx.GetDbMgr().IsBlockExist(block.BlockId, true, user.nodename)
	if err != nil {
		return err
	}

	if isCached {
		molaproducer_log.Debugf("<%s> cached block, update block", user.groupId)
	}

	//Save block to cache
	err = nodectx.GetDbMgr().AddBlock(block, true, user.nodename)
	if err != nil {
		return err
	}

	//check if parent of block exist
	parentExist, err := nodectx.GetDbMgr().IsParentExist(block.PrevBlockId, false, user.nodename)
	if err != nil {
		return err
	}

	if !parentExist {
		molauser_log.Debugf("<%s> parent of block <%s> is not exist", user.groupId, block.BlockId)
		return errors.New("PARENT_NOT_EXIST")
	}

	//get parent block
	parentBlock, err := nodectx.GetDbMgr().GetBlock(block.PrevBlockId, false, user.nodename)
	if err != nil {
		return err
	}

	//valid block with parent block
	valid, err := IsBlockValid(block, parentBlock)
	if !valid {
		molauser_log.Debugf("<%s> remove invalid block <%s> from cache", user.groupId, block.BlockId)
		molauser_log.Warningf("<%s> invalid block <%s>", user.groupId, err.Error())
		return nodectx.GetDbMgr().RmBlock(block.BlockId, true, user.nodename)
	}

	//search cache, gather all blocks can be connected with this block
	blocks, err := nodectx.GetDbMgr().GatherBlocksFromCache(block, true, user.nodename)
	if err != nil {
		return err
	}

	//get all trxs from those blocks
	var trxs []*chestnutpb.Trx
	trxs, err = GetAllTrxs(blocks)
	if err != nil {
		return err
	}

	//apply those trxs
	err = user.applyTrxs(trxs, user.nodename)
	if err != nil {
		return err
	}

	//move gathered blocks from cache to chain
	for _, block := range blocks {
		molauser_log.Debugf("<%s> move block <%s> from cache to chain", user.groupId, block.BlockId)
		err := nodectx.GetDbMgr().AddBlock(block, false, user.nodename)
		if err != nil {
			return err
		}

		err = nodectx.GetDbMgr().RmBlock(block.BlockId, true, user.nodename)
		if err != nil {
			return err
		}
	}

	//update block produced count
	for _, block := range blocks {
		err := nodectx.GetDbMgr().AddProducedBlockCount(user.groupId, block.ProducerPubKey, user.nodename)
		if err != nil {
			return err
		}
	}

	//calculate new height
	molauser_log.Debugf("<%s> height before recal <%d>", user.groupId, user.grpItem.HighestHeight)
	topBlock, err := nodectx.GetDbMgr().GetBlock(user.grpItem.HighestBlockId, false, user.nodename)
	if err != nil {
		return err
	}
	newHeight, newHighestBlockId, err := RecalChainHeight(blocks, user.grpItem.HighestHeight, topBlock, user.nodename)
	if err != nil {
		return err
	}
	molauser_log.Debugf("<%s> new height <%d>, new highest blockId %v", user.groupId, newHeight, newHighestBlockId)

	//if the new block is not highest block after recalculate, we need to "trim" the chain
	if newHeight < user.grpItem.HighestHeight {

		//from parent of the new blocks, get all blocks not belong to the longest path
		resendBlocks, err := GetTrimedBlocks(blocks, user.nodename)
		if err != nil {
			return err
		}

		var resendTrxs []*chestnutpb.Trx
		resendTrxs, err = GetMyTrxs(resendBlocks, user.nodename, user.grpItem.UserSignPubkey)

		if err != nil {
			return err
		}

		UpdateResendCount(resendTrxs)
		err = user.resendTrx(resendTrxs)
	}

	return user.cIface.UpdChainInfo(newHeight, newHighestBlockId)
}

//resend all trx in the list
func (user *MolassesUser) resendTrx(trxs []*chestnutpb.Trx) error {
	molauser_log.Debugf("<%s> resendTrx called", user.groupId)
	for _, trx := range trxs {
		molauser_log.Debugf("<%s> resend Trx <%s>", user.groupId, trx.TrxId)
		user.cIface.GetProducerTrxMgr().ResendTrx(trx)
	}
	return nil
}

func (user *MolassesUser) applyTrxs(trxs []*chestnutpb.Trx, nodename string) error {
	molauser_log.Debugf("<%s> applyTrxs called", user.groupId)
	for _, trx := range trxs {
		//check if trx already applied
		isExist, err := nodectx.GetDbMgr().IsTrxExist(trx.TrxId, nodename)
		if err != nil {
			molauser_log.Debugf("<%s> %s", user.groupId, err.Error())
			continue
		}

		if isExist {
			molauser_log.Debugf("<%s> trx <%s> existed, update trx only", user.groupId, trx.TrxId)
			nodectx.GetDbMgr().AddTrx(trx)
			continue
		}

		originalData := trx.Data

		//new trx, apply it
		if trx.Type == chestnutpb.TrxType_POST && user.grpItem.EncryptType == chestnutpb.GroupEncryptType_PRIVATE {
			//for post, private group, encrypted by pgp for all announced group user
			ks := localcrypto.GetKeystore()
			decryptData, err := ks.Decrypt(user.grpItem.UserEncryptPubkey, trx.Data)
			if err != nil {
				return err
			}

			//set trx.Data to decrypted []byte
			trx.Data = decryptData
		} else {
			//decode trx data
			ciperKey, err := hex.DecodeString(user.grpItem.CipherKey)
			if err != nil {
				return err
			}

			decryptData, err := localcrypto.AesDecode(trx.Data, ciperKey)
			if err != nil {
				return err
			}

			//set trx.Data to decrypted []byte
			trx.Data = decryptData
		}

		molauser_log.Debugf("<%s> try apply trx <%s>", user.groupId, trx.TrxId)
		//apply trx content
		switch trx.Type {
		case chestnutpb.TrxType_POST:
			molauser_log.Debugf("<%s> apply POST trx", user.groupId)
			nodectx.GetDbMgr().AddPost(trx, nodename)
		case chestnutpb.TrxType_AUTH:
			molauser_log.Debugf("<%s> apply AUTH trx", user.groupId)
			nodectx.GetDbMgr().UpdateBlkListItem(trx, nodename)
		case chestnutpb.TrxType_PRODUCER:
			molauser_log.Debugf("<%s> apply PRODUCER trx", user.groupId)
			nodectx.GetDbMgr().UpdateProducer(trx, nodename)
			user.cIface.UpdProducerList()
			user.cIface.CreateConsensus()
		case chestnutpb.TrxType_ANNOUNCE:
			molauser_log.Debugf("<%s> apply ANNOUNCE trx", user.groupId)
			nodectx.GetDbMgr().UpdateAnnounce(trx, nodename)
		case chestnutpb.TrxType_SCHEMA:
			molauser_log.Debugf("<%s> apply SCHEMA trx", user.groupId)
			nodectx.GetDbMgr().UpdateSchema(trx, nodename)
		default:
			molauser_log.Warningf("<%s> unsupported msgType <%s>", user.groupId, trx.Type)
		}

		//set trx data to original(encrypted)
		trx.Data = originalData

		//save trx to db
		nodectx.GetDbMgr().AddTrx(trx, nodename)
	}

	return nil
}



