// Package chain provides chain for chestnut.
package chain

import (
	"bytes"
	"errors"
	"time"

	guuid "github.com/google/uuid"
	p2pcrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/lixvyang/chestnut/nodectx"
	chestnutpb "github.com/lixvyang/chestnut/pb"
	"google.golang.org/protobuf/proto"
)


func CreateBlock(oldBlock *chestnutpb.Block, trxs []*chestnutpb.Trx, groupPublicKey []byte, opts ...string) (*chestnutpb.Block, error) {
	var newBlock chestnutpb.Block
	blockId := guuid.New()

	newBlock.BlockId = blockId.String()
	newBlock.GroupId = oldBlock.GroupId
	newBlock.PrevBlockId = oldBlock.BlockId
	newBlock.PreviousHash = oldBlock.Hash
	for _, trx := range trxs {
		trxclone := &chestnutpb.Trx{}

		clonedtrxbuff, err := proto.Marshal(trx)

		err = proto.Unmarshal(clonedtrxbuff, trxclone)
		if err != nil {
			return nil, err
		}
		newBlock.Trxs = append(newBlock.Trxs, trxclone)
	}
	newBlock.ProducerPubKey = p2pcrypto.ConfigEncodeKey(groupPublicKey)
	newBlock.TimeStamp = time.Now().UnixNano()

	bbytes, err := proto.Marshal(&newBlock)
	if err != nil {
		return nil, err
	}

	hash := Hash(bbytes)
	newBlock.Hash = hash

	signature, err := nodectx.GetNodeCtx().Keystore.SignByKeyName(newBlock.GroupId,hash,opts...)
	if err != nil {
		return nil, err
	}

	newBlock.Signature = signature
	return &newBlock,nil
}


func CreateGeneisBlock(groupId string, groupPublicKey p2pcrypto.PubKey) (*chestnutpb.Block, error)  {
	encodedgroupPubkey, err := p2pcrypto.MarshalPublicKey(groupPublicKey)
	if err != nil {
		return nil, err
	}

	var genesisBlock chestnutpb.Block
	genesisBlock.BlockId = guuid.New().String()
	genesisBlock.GroupId = groupId
	genesisBlock.PrevBlockId = ""
	genesisBlock.PreviousHash = nil
	genesisBlock.TimeStamp = time.Now().UnixNano()

	genesisBlock.ProducerPubKey = p2pcrypto.ConfigEncodeKey(encodedgroupPubkey)
	genesisBlock.Trxs = nil
	
	bbytes, err := proto.Marshal(&genesisBlock)
	if err != nil {
		return nil, err
	}

	hash := Hash(bbytes)
	genesisBlock.Hash = hash

	signature, err := nodectx.GetNodeCtx().Keystore.SignByKeyName(genesisBlock.GroupId, hash)
	if err != nil {
		return nil, err
	}

	genesisBlock.Signature = signature
	return &genesisBlock, nil
}


func IsBlockValid(newBlock, oldBlock *chestnutpb.Block) (bool, error) {
	// deep copy newBlock by the protobuf. chestnutpb.Block is a protopbbuf defined struct.
	clonedblockbuff, err := proto.Marshal(newBlock)
	if err != nil {
		return false, err
	}

	var blockWithoutHash *chestnutpb.Block
	blockWithoutHash = &chestnutpb.Block{}

	err = proto.Unmarshal(clonedblockbuff,blockWithoutHash)
	if err != nil {
		return false, err
	}

	// set hash to ""
	blockWithoutHash.Hash = nil
	blockWithoutHash = &chestnutpb.Block{}

	bbytes, err := proto.Marshal(blockWithoutHash)
	if err != nil {
		return  false, err
	}
	hash := Hash(bbytes)
	if res := bytes.Compare(hash, newBlock.Hash); res != 0 {
		return false, errors.New("Hash for new block is invalid")
	}

	if res := bytes.Compare(newBlock.PreviousHash, oldBlock.Hash); res != 0 {
		return false, errors.New("PreviousHash mismatch")
	}

	if newBlock.PrevBlockId != oldBlock.BlockId {
		return false, errors.New("Previous BlockId mismatch")
	}

	// create pubkey
	serializedpub, err := p2pcrypto.ConfigDecodeKey(newBlock.ProducerPubKey)
	if err != nil {
		return false, err
	}

	pubkey, err := p2pcrypto.UnmarshalPublicKey(serializedpub)
	if err != nil {
		return false, err
	}

	verify, err := pubkey.Verify(newBlock.Hash,newBlock.Signature)
	return verify, err

}

