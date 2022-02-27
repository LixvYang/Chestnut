// Package chain provides chain for chestnut.
package chain

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	guuid "github.com/google/uuid"
	logging "github.com/ipfs/go-log/v2"
	p2pcrypto "github.com/libp2p/go-libp2p-core/crypto"
	localcrypto "github.com/lixvyang/chestnut/crypto"
	"github.com/lixvyang/chestnut/nodectx"
	chestnutpb "github.com/lixvyang/chestnut/pb"
	"github.com/lixvyang/chestnut/pubsubconn"
	"google.golang.org/protobuf/proto"
)

const (
	Hours = 0
	Mins  = 5
	Sec   = 0
)

const OBJECT_SIZE_LIMIT = 200 * 1024 //(200Kb)

var trxmgr_log = logging.Logger("trxmgr")

type TrxMgr struct {
	nodename string
	groupItem *chestnutpb.GroupItem
	psconn pubsubconn.PubSubConn
	groupId string
}

func (trxMgr *TrxMgr) Init(groupItem *chestnutpb.GroupItem, psconn pubsubconn.PubSubConn) {
	trxMgr.groupItem = groupItem
	trxMgr.psconn = psconn
	trxMgr.groupId = groupItem.GroupId
}

func (trxMgr *TrxMgr) SetNodeName(nodename string) {
	trxMgr.nodename = nodename
}

func (trxMgr *TrxMgr) CreateTrxWithoutSign(msgType chestnutpb.TrxType, data []byte) (*chestnutpb.Trx, []byte, error) {
	var trx chestnutpb.Trx

	trxId := guuid.New()
	trx.TrxId = trxId.String()
	trx.Type = msgType
	trx.GroupId = trxMgr.groupItem.GroupId
	trx.SenderPubkey = trxMgr.groupItem.UserSignPubkey

	var encryptdData []byte

	if msgType == chestnutpb.TrxType_POST && trxMgr.groupItem.EncryptType == chestnutpb.GroupEncryptType_PRIVATE {
		//for post, private group, encrypted by age for all announced group users
		var err error
		announcedUser, err := nodectx.GetDbMgr().GetAnnouncedUsersByGroup(trxMgr.groupItem.GroupId)

		var pubkeys []string
		for _, item := range announcedUser {
			if item.Result == chestnutpb.ApproveType_APPROVED {
				pubkeys = append(pubkeys, item.EncryptPubkey)
			}
		}

		ks := localcrypto.GetKeystore()
		encryptdData, err = ks.EncryptTo(pubkeys, data)
		if err != nil {
			return &trx, []byte(""), err
		}
	} else {
		var err error
		ciperKey, err := hex.DecodeString(trxMgr.groupItem.CipherKey)
		if err != nil {
			return &trx, []byte(""), err
		}
		encryptdData, err = localcrypto.AesEncrypt(data, ciperKey)
		if err != nil {
			return &trx, []byte(""), err
		}
	}
	trx.Data = encryptdData

	trx.TimeStamp = time.Now().UnixNano()
	trx.Version = nodectx.GetNodeCtx().Version
	timein := time.Now().Local().Add(time.Hour*time.Duration(Hours) +
		time.Minute*time.Duration(Mins) +
		time.Second*time.Duration(Sec))
	trx.Expired = timein.UnixNano()

	bytes, err := proto.Marshal(&trx)
	if err != nil {
		return &trx, []byte(""), err
	}
	hashed := localcrypto.Hash(bytes)
	return &trx, hashed, nil
}

func (trxMgr *TrxMgr) CreateTrx(msgType chestnutpb.TrxType, data []byte) (*chestnutpb.Trx,  error)  {
	trx, hashed, err := trxMgr.CreateTrxWithoutSign(msgType, data)
	if err != nil {
		return trx, err
	}	

	ks := nodectx.GetNodeCtx().Keystore
	keyname := trxMgr.groupItem.GroupId
	if trxMgr.nodename != "" {
		keyname = fmt.Sprintf("%s_%s", trxMgr.nodename, trxMgr.groupItem.GroupId)
	}
	signature, err := ks.SignByKeyName(keyname, hashed)
	if err != nil {
		return trx, err
	}
	trx.SenderSign = signature
	return trx, nil
}


func (trxMgr *TrxMgr) VerifyTrx(trx *chestnutpb.Trx) (bool, error) {
	//clone trxMsg to verify
	clonetrxmsg := &chestnutpb.Trx{
		TrxId:        trx.TrxId,
		Type:         trx.Type,
		GroupId:      trx.GroupId,
		SenderPubkey: trx.SenderPubkey,
		Data:         trx.Data,
		TimeStamp:    trx.TimeStamp,
		Version:      trx.Version,
		Expired:      trx.Expired}

	bytes, err := proto.Marshal(clonetrxmsg)
	if err != nil {
		return false, err
	}

	hashed := localcrypto.Hash(bytes)

	//create pubkey
	serializedpub, err := p2pcrypto.ConfigDecodeKey(trx.SenderPubkey)
	if err != nil {
		return false, err
	}

	pubkey, err := p2pcrypto.UnmarshalPublicKey(serializedpub)
	if err != nil {
		return false, err
	}

	verify, err := pubkey.Verify(hashed, trx.SenderSign)
	return verify, err
}

func (trxMgr *TrxMgr) SendUpdAuthTrx(item *chestnutpb.DenyUserItem) (string, error) {
	trxmgr_log.Debugf("<%s> SendUpdAuthTrx called", trxMgr.groupId)

	encodedcontent, err := proto.Marshal(item)
	if err != nil {
		return "", err
	}

	trx, err := trxMgr.CreateTrx(chestnutpb.TrxType_AUTH, encodedcontent)
	err = trxMgr.sendTrx(trx)
	if err != nil {
		return "INVALID_TRX", err
	}

	return trx.TrxId, nil
}

func (trxMgr *TrxMgr) SendRegProducerTrx(item *chestnutpb.ProducerItem) (string, error) {
	trxmgr_log.Debugf("<%s> SendRegProducerTrx called", trxMgr.groupId)
	encodedcontent, err := proto.Marshal(item)
	if err != nil {
		return "", err
	}
	trx, err := trxMgr.CreateTrx(chestnutpb.TrxType_PRODUCER, encodedcontent)
	err = trxMgr.sendTrx(trx)
	if err != nil {
		return "INVALID_TRX", err
	}

	return trx.TrxId, nil
}


func (trxMgr *TrxMgr) SendAnnounceTrx(item *chestnutpb.AnnounceItem) (string, error) {
	trxmgr_log.Debugf("<%s> SendAnnounceTrx called", trxMgr.groupId)
	encodedcontent, err := proto.Marshal(item)
	if err != nil {
		return "", err
	}

	trx, err := trxMgr.CreateTrx(chestnutpb.TrxType_ANNOUNCE, encodedcontent)
	err = trxMgr.sendTrx(trx)
	if err != nil {
		return "INVALID_TRX", err
	}

	return trx.TrxId, nil
}

func (trxMgr *TrxMgr) SendUpdSchemaTrx(item *chestnutpb.SchemaItem) (string, error) {
	trxmgr_log.Debugf("<%s> SendUpdSchemaTrx called", trxMgr.groupId)
	encodedcontent, err := proto.Marshal(item)
	if err != nil {
		return "", err
	}

	trx, err := trxMgr.CreateTrx(chestnutpb.TrxType_SCHEMA, encodedcontent)
	err = trxMgr.sendTrx(trx)
	if err != nil {
		return "INVALID_TRX", err
	}

	return trx.TrxId, nil
}


func (trxMgr *TrxMgr) SendReqBlockResp(req *chestnutpb.ReqBlock, block *chestnutpb.Block, result chestnutpb.ReqBlkResult) error {
	trxmgr_log.Debugf("<%s> SendReqBlockResp called", trxMgr.groupId)

	var reqBlockRespItem chestnutpb.ReqBlockResp
	reqBlockRespItem.Result = result
	reqBlockRespItem.ProviderPubkey = trxMgr.groupItem.UserSignPubkey
	reqBlockRespItem.RequesterPubkey = req.UserId
	reqBlockRespItem.GroupId = req.GroupId
	reqBlockRespItem.BlockId = req.BlockId

	pbBytesBlock, err := proto.Marshal(block)
	if err != nil {
		return err
	}
	reqBlockRespItem.Block = pbBytesBlock

	bItemBytes, err := proto.Marshal(&reqBlockRespItem)
	if err != nil {
		return err
	}

	//send ask next block trx out
	trx, err := trxMgr.CreateTrx(chestnutpb.TrxType_REQ_BLOCK_RESP, bItemBytes)
	if err != nil {
		trxmgr_log.Warningf(err.Error())
		return err
	}

	return trxMgr.sendTrx(trx)
}


func (trxMgr *TrxMgr) SendReqBlockForward(block *chestnutpb.Block) error {
	trxmgr_log.Debugf("<%s> SendReqBlockForward called", trxMgr.groupId)

	var reqBlockItem chestnutpb.ReqBlock
	reqBlockItem.BlockId = block.BlockId
	reqBlockItem.GroupId = block.GroupId
	reqBlockItem.UserId = trxMgr.groupItem.UserSignPubkey

	bItemBytes, err := proto.Marshal(&reqBlockItem)
	if err != nil {
		group_log.Warningf(err.Error())
		return err
	}

	trx, err := trxMgr.CreateTrx(chestnutpb.TrxType_REQ_BLOCK_FORWARD, bItemBytes)
	if err != nil {
		group_log.Warningf(err.Error())
		return err
	}

	return trxMgr.sendTrx(trx)
}

func (trxMgr *TrxMgr) SendReqBlockBackward(block *chestnutpb.Block) error {
	trxmgr_log.Debugf("<%s> SendReqBlockBackward called", trxMgr.groupId)

	var reqBlockItem chestnutpb.ReqBlock
	reqBlockItem.BlockId = block.BlockId
	reqBlockItem.GroupId = block.GroupId
	reqBlockItem.UserId = trxMgr.groupItem.UserSignPubkey

	bItemBytes, err := proto.Marshal(&reqBlockItem)
	if err != nil {
		group_log.Warningf(err.Error())
		return err
	}

	trx, err := trxMgr.CreateTrx(chestnutpb.TrxType_REQ_BLOCK_BACKWARD, bItemBytes)
	if err != nil {
		group_log.Warningf(err.Error())
		return err
	}

	return trxMgr.sendTrx(trx)
}

func (trxMgr *TrxMgr) SendBlockProduced(blk *chestnutpb.Block) error {
	trxmgr_log.Debugf("<%s> SendBlockProduced called", trxMgr.groupId)
	encodedcontent, err := proto.Marshal(blk)
	if err != nil {
		return err
	}
	trx, err := trxMgr.CreateTrx(chestnutpb.TrxType_BLOCK_PRODUCED, encodedcontent)
	if err != nil {
		return err
	}
	return trxMgr.sendTrx(trx)
}

func (trxMgr *TrxMgr) PostBytes(trxtype chestnutpb.TrxType, encodedcontent []byte) (string, error) {
	trxmgr_log.Debugf("<%s> PostBytes called", trxMgr.groupId)
	trx, err := trxMgr.CreateTrx(trxtype, encodedcontent)
	err = trxMgr.sendTrx(trx)
	if err != nil {
		return "INVALID_TRX", err
	}

	return trx.TrxId, nil
}

func (trxMgr *TrxMgr) PostAny(content proto.Message) (string, error) {
	trxmgr_log.Debugf("<%s> PostAny called", trxMgr.groupId)

	encodedcontent, err := chestnutpb.ContentToBytes(content)
	if err != nil {
		return "", err
	}

	trxmgr_log.Debugf("<%s> content size <%d>", trxMgr.groupId, binary.Size(encodedcontent))
	if binary.Size(encodedcontent) > OBJECT_SIZE_LIMIT {
		err := errors.New("Content size over 200Kb")
		return "", err
	}

	return trxMgr.PostBytes(chestnutpb.TrxType_POST, encodedcontent)
}

func (trxMgr *TrxMgr) ResendTrx(trx *chestnutpb.Trx) error {
	trxmgr_log.Debugf("<%s> ResendTrx called", trxMgr.groupId)
	return trxMgr.sendTrx(trx)
}

func (trxMgr *TrxMgr) CustomSendTrx(trx *chestnutpb.Trx) error {
	trxmgr_log.Debugf("<%s> CustomSendTrx called", trxMgr.groupId)
	return trxMgr.sendTrx(trx)
}

func (trxMgr *TrxMgr) SendBlock(blk *chestnutpb.Block) error {
	trxmgr_log.Debugf("<%s> SendBlock called", trxMgr.groupId)

	var pkg *chestnutpb.Package
	pkg = &chestnutpb.Package{}

	pbBytes, err := proto.Marshal(blk)
	if err != nil {
		return err
	}

	pkg.Type = chestnutpb.PackageType_BLOCK
	pkg.Data = pbBytes

	pkgBytes, err := proto.Marshal(pkg)
	if err != nil {
		return err
	}

	return trxMgr.psconn.Publish(pkgBytes)
}

func (trxMgr *TrxMgr) sendTrx(trx *chestnutpb.Trx) error {
	trxmgr_log.Debugf("<%s> sendTrx called", trxMgr.groupId)
	var pkg *chestnutpb.Package
	pkg = &chestnutpb.Package{}

	pbBytes, err := proto.Marshal(trx)
	if err != nil {
		return err
	}

	pkg.Type = chestnutpb.PackageType_TRX
	pkg.Data = pbBytes

	pkgBytes, err := proto.Marshal(pkg)
	if err != nil {
		return err
	}

	return trxMgr.psconn.Publish(pkgBytes)
}
