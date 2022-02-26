// Package storage provides storage for chestnut.
package storage

import (
	"errors"

	logging "github.com/ipfs/go-log/v2"
	chestnutpb "github.com/lixvyang/chestnut/pb"
	"google.golang.org/protobuf/proto"
)

var dbmgr_log = logging.Logger("dbmgr")

const TRX_PREFIX = "trx" //trx
const BLK_PREFIX = "blk" //block
const GRP_PREFIX = "grp" //group
const CNT_PREFIX = "cnt" //content
const ATH_PREFIX = "ath" //auth
const PRD_PREFIX = "prd" //producer
const ANN_PREFIX = "ann" //announce
const SMA_PREFIX = "sma" //schema
const CHD_PREFIX = "chd" //cached

type DbMgr struct {
	GroupInfoDb ChestnutStorage
	Db	ChestnutStorage
	Auth ChestnutStorage
	DataPath string
}

func (dbMgr *DbMgr) CloseDb() {
	dbMgr.GroupInfoDb.Close()
	dbMgr.Db.Close()
	dbmgr_log.Infof("ChainCtx Db Closed")
}

func (dbMgr *DbMgr) TryMigration(nodeDataVer int)  {
	if nodeDataVer == 0 {
		dbmgr_log.Info("Migration v0")
		groupItemsBytes, err := dbMgr.GetGroupsBytes()
		if err != nil {
			for _, b := range groupItemsBytes {
				var item *chestnutpb.GroupItem
				item = &chestnutpb.GroupItem{}
				proto.Unmarshal(b, item)
				if item.CipherKey == "" {
					itemv0 := &chestnutpb.GroupItemV0{}
					proto.Unmarshal(b, itemv0)
					if itemv0.CipherKey != "" { //ok
						item.LastUpdate = itemv0.LastUpdate
						item.HighestHeight = itemv0.HighestHeight
						item.HighestBlockId = itemv0.HighestBlockId
						item.GenesisBlock = itemv0.GenesisBlock
						item.EncryptType = itemv0.EncryptType
						item.ConsenseType = itemv0.ConsenseType
						item.CipherKey = itemv0.CipherKey
						item.AppKey = itemv0.AppKey
						//add group to db
						value, err := proto.Marshal(item)
						if err == nil {
							dbMgr.GroupInfoDb.Set([]byte(item.GroupId), value)
							dbmgr_log.Infof("db migration v0 for group %s", item.GroupId)
						}
					}
				}
			}
		}
	}
}

// save trx
func (dbMgr *DbMgr) AddTrx(trx *chestnutpb.Trx, prefix ...string) error {
	nodeprefix := getPrefix(prefix...)
	key := nodeprefix + TRX_PREFIX + "_" + trx.TrxId
	value, err := proto.Marshal(trx)
	if err != nil {
		return err
	}
	return dbMgr.Db.Set([]byte(key), value)
}

// UNUSED
// Remove Trx
func (dbMgr *DbMgr) RmTrx(trxId string, prefix ...string) error {
	nodeprefix := getPrefix(prefix...)
	key := nodeprefix + TRX_PREFIX + "_" + trxId
	return dbMgr.Db.Delete([]byte(key))
}

// get trx
func (dbMgr *DbMgr) GetTrx(trxId string, prefix ...string) (*chestnutpb.Trx, error) {
	nodeprefix := getPrefix(prefix...)
	key := nodeprefix + TRX_PREFIX + "_" + trxId
	value, err := dbMgr.Db.Get([]byte(key))
	if err != nil {
		return nil, err
	}
	
	var trx chestnutpb.Trx
	err = proto.Unmarshal(value, &trx)
	if err != nil {
		return nil, err
	}
	return &trx, nil
}

func (dbMgr *DbMgr) UpdTrx(trx *chestnutpb.Trx, prefix ...string) error {
	return dbMgr.AddTrx(trx, prefix...)
}

func (dbMgr *DbMgr) IsTrxExist(trxId string, prefix ...string) (bool, error) {
	nodeprefix := getPrefix(prefix...)
	key := nodeprefix + TRX_PREFIX + "_" + trxId
	return dbMgr.Db.IsExist([]byte(key))
}

func (dbMgr *DbMgr) AddGensisBlock(gensisBlock *chestnutpb.Block, prefix ...string) error {
	nodePrefix := getPrefix(prefix...)
	key := nodePrefix + BLK_PREFIX + "_" + gensisBlock.BlockId
	
	chunk := chestnutpb.BlockDbChunk{}
	chunk.BlockId = gensisBlock.BlockId
	chunk.BlockItem = gensisBlock
	value, err := proto.Marshal(&chunk)
	if err != nil {
		return err
	}
	return dbMgr.Db.Set([]byte(key), value)
}

// check if block existed
func (dbMgr *DbMgr) IsBlockExist(blockId string, cached bool, prefix ...string) (bool, error) {
	nodeprefix := getPrefix(prefix...)
	var key string 
	if cached {
		key = nodeprefix + CHD_PREFIX + "_" + BLK_PREFIX + "_" + blockId
	} else {
		key = nodeprefix + BLK_PREFIX + "_" + blockId
	}
	return dbMgr.Db.IsExist([]byte(key))
}

// check if parent block existed
func (dbMgr *DbMgr) IsParentExist(parentBlockId string, cached bool, prefix ...string) (bool, error) {
	nodeprefix := getPrefix(prefix...)
	var pKey string 
	if cached {
		pKey = nodeprefix + CHD_PREFIX + "_" + BLK_PREFIX + "_" + parentBlockId
	} else {
		pKey = nodeprefix + BLK_PREFIX + "_" + parentBlockId
	}

	return dbMgr.Db.IsExist([]byte(pKey))
}

// add Block
func (dbMgr *DbMgr) AddBlock(newBlock *chestnutpb.Block, cached bool, prefix ...string) error {
	// create new chunk
	var chunk *chestnutpb.BlockDbChunk
	chunk = &chestnutpb.BlockDbChunk{}
	chunk.BlockId = newBlock.BlockId
	chunk.BlockItem = newBlock

	if cached {
		chunk.Height = -1
		chunk.ParentBlockId = ""
	} else {
		// try get parent chunk
		pChunk, err := dbMgr.getBlockChunk(newBlock.PrevBlockId, cached, prefix...)
		if err != nil {
			return err
		}

		// update parent chunk
		pChunk.SubBlockId = append(pChunk.SubBlockId, newBlock.BlockId)
		err = dbMgr.saveBlockChunk(pChunk, cached, prefix...)
		if err != nil {
			return err
		}

		chunk.Height = pChunk.Height + 1
		chunk.ParentBlockId = pChunk.BlockId
	}
	// save chunk
	return dbMgr.saveBlockChunk(chunk, cached, prefix...)
}

//remove block
func (dbMgr *DbMgr) RmBlock(blockId string, cached bool, prefix ...string) error {
	nodeprefix := getPrefix(prefix...)
	var key string
	if cached {
		key = nodeprefix + CHD_PREFIX + "_" + BLK_PREFIX + "_" + blockId
	} else {
		key = nodeprefix + BLK_PREFIX + "_" + blockId
	}

	return dbMgr.Db.Delete([]byte(key))
}

//get block by block_id
func (dbMgr *DbMgr) GetBlock(blockId string, cached bool, prefix ...string) (*chestnutpb.Block, error) {
	pChunk, err := dbMgr.getBlockChunk(blockId, cached, prefix...)
	if err != nil {
		return nil, err
	}
	return pChunk.BlockItem, nil
}

func (dbMgr *DbMgr) GatherBlocksFromCache(newBlock *chestnutpb.Block, cached bool, prefix ...string) ([]*chestnutpb.Block, error) {
	nodeprefix := getPrefix(prefix...)
	var blocks []*chestnutpb.Block
	blocks = append(blocks, newBlock)
	pointer1 := 0 // point to head
	pointer2 := 0 // point to tail

	pre := nodeprefix + CHD_PREFIX + "_" + BLK_PREFIX + "_"
	for {
		err := dbMgr.Db.PrefixForeach([]byte(pre), func(k, v []byte, err error)error {
			if err != nil {
				return err
			}

			chunk := chestnutpb.BlockDbChunk{}
			perr := proto.Unmarshal(v, &chunk)
			if perr != nil {
				return perr
			}

			if chunk.BlockItem.PrevBlockId == blocks[pointer1].BlockId {
				blocks = append(blocks, chunk.BlockItem)
				pointer2++
			}
			return nil
		})
		if err != nil {
			return blocks, err
		}

		if pointer1 == pointer2{
			break
		}

		pointer1++
	}
	return blocks, nil
}

func (dbMgr *DbMgr) GetBlockHeight(blockId string, prefix ...string) (int64, error) {
	pChunk, err := dbMgr.getBlockChunk(blockId, false, prefix...)
	if err != nil {
		return -1, err
	}
	return pChunk.Height, nil
}

func (dbMgr *DbMgr) GetSubBlock(blockId string, prefix ...string) ([]*chestnutpb.Block, error) {
	var result []*chestnutpb.Block
	chunk, err := dbMgr.getBlockChunk(blockId, false, prefix...)
	if err != nil {
		return nil, err
	}
	for _, subChunk := range chunk.SubBlockId {
		subChunk, err := dbMgr.getBlockChunk(subChunk, false, prefix...)
		if err != nil {
			return nil, err
		}
		result = append(result, subChunk.BlockItem)
	}
	return result, nil
}

func (dbMgr *DbMgr) GetParentBlock(blockId string, prefix ...string) (*chestnutpb.Block, error) {
	chunk, err := dbMgr.getBlockChunk(blockId, false, prefix...)
	if err != nil {
		return nil, err
	}
	parentChunk, err := dbMgr.getBlockChunk(chunk.ParentBlockId, false, prefix...)
	return parentChunk.BlockItem, err
}


func (dbMgr *DbMgr) getBlockChunk(blockId string, cached bool, prefix ...string) (*chestnutpb.BlockDbChunk, error) {
	nodeprefix := getPrefix(prefix...)
	var key string
	if cached {
		key = nodeprefix + CHD_PREFIX + "_" + BLK_PREFIX + "_" + blockId
	} else {
		key = nodeprefix + BLK_PREFIX + "_" + blockId
	}
	pChunk := chestnutpb.BlockDbChunk{}
	value, err := dbMgr.Db.Get([]byte(key))
	if err != nil {
		return nil, err
	}

	err = proto.Unmarshal(value, &pChunk)
	if err != nil {
		return nil, err
	}
	return &pChunk, err
}

// save block chunk
func (dbMgr *DbMgr) saveBlockChunk(chunk *chestnutpb.BlockDbChunk, cached bool, prefix ...string) error {
	nodeprefix := getPrefix(prefix...)
	var key string
	if cached {
		key = nodeprefix + CHD_PREFIX + "_" + BLK_PREFIX + "_" + chunk.BlockId
	} else {
		key = nodeprefix + BLK_PREFIX + "_" + chunk.BlockId
	}

	value, err := proto.Marshal(chunk)
	if err != nil {
		return err
	}
	return dbMgr.Db.Set([]byte(key), value)
}

// add group
func (dbMgr *DbMgr) AddGroup(groupItem *chestnutpb.GroupItem) error {
	//check if group exist
	exist, err := dbMgr.GroupInfoDb.IsExist([]byte(groupItem.GroupId))
	if exist {
		return errors.New("Group with same GroupId existed")
	}

	//add group to db
	value, err := proto.Marshal(groupItem)
	if err != nil {
		return err
	}
	return dbMgr.GroupInfoDb.Set([]byte(groupItem.GroupId), value)
}




// Get group list
func (dbMgr *DbMgr) GetGroupsBytes() ([][]byte, error) {
	var groupItemList [][]byte
	// test only, show db contents
	err := dbMgr.GroupInfoDb.Foreach(func(k, v []byte, err error) error {
		if err != nil {
			return err
		}

		groupItemList = append(groupItemList, v)
		return nil
	})
	return groupItemList, err
}

func getPrefix(prefix ...string) string {
	nodeprefix := ""
	if len(prefix) == 1 {
		nodeprefix = prefix[0] + "_"
	}
	return nodeprefix
}