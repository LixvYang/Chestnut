// Package chain provides chain for chestnut.
package chain

import (
	"bytes"

	"github.com/lixvyang/chestnut/nodectx"

	logging "github.com/ipfs/go-log/v2"

	chestnutpb "github.com/lixvyang/chestnut/pb"

	localcrypto "github.com/lixvyang/chestnut/crypto"
)

var molautil_log = logging.Logger("util")

func Hash(data []byte) []byte {
	return localcrypto.Hash(data)
}

//find the highest block from the block tree
func RecalChainHeight(blocks []*chestnutpb.Block, currentHeight int64, currentHighestBlock *chestnutpb.Block, nodename string) (int64, string, error) {
	molautil_log.Debug("RecalChainHeight called")

	newHighestHeight := currentHeight
	newHighestBlockId := currentHighestBlock.BlockId
	newHighestBlock := currentHighestBlock

	for _, block := range blocks {
		blockHeight, err := nodectx.GetDbMgr().GetBlockHeight(block.BlockId, nodename)
		if err != nil {
			return -1, "INVALID_BLOCK_ID", err
		}
		if blockHeight > newHighestHeight {
			newHighestHeight = blockHeight
			newHighestBlockId = block.BlockId
			newHighestBlock = block
		} else if blockHeight == newHighestHeight {
			//comparing two hash bytes lexicographicall
			if bytes.Compare(newHighestBlock.Hash[:], block.Hash[:]) == -1 { //-1 means ohash < nhash, and we want keep the larger one
				newHighestHeight = blockHeight
				newHighestBlockId = block.BlockId
				newHighestBlock = block
			}
		}
	}

	return newHighestHeight, newHighestBlockId, nil
}

//from root of the new block tree, get all blocks trimed when not belong to longest path
func GetTrimedBlocks(blocks []*chestnutpb.Block, nodename string) ([]string, error) {
	molautil_log.Debug("GetTrimedBlocks called")
	var cache map[string]bool
	var longestPath []string
	var result []string

	cache = make(map[string]bool)

	err := dfs(blocks, cache, longestPath, nodename)

	for _, blockId := range longestPath {
		if _, ok := cache[blockId]; !ok {
			result = append(result, blockId)
		}
	}

	return result, err
}

func dfs(blocks []*chestnutpb.Block, cache map[string]bool, result []string, nodename string) error {
	molautil_log.Debug("dfs called")
	for _, block := range blocks {
		if _, ok := cache[block.BlockId]; !ok {
			cache[block.BlockId] = true
			result = append(result, block.BlockId)
			subBlocks, err := nodectx.GetDbMgr().GetSubBlock(block.BlockId, nodename)
			if err != nil {
				return err
			}
			err = dfs(subBlocks, cache, result, nodename)
		}
	}
	return nil
}


//get all trx belongs to me from the block list
func GetMyTrxs(blockIds []string, nodename string, userSignPubkey string) ([]*chestnutpb.Trx, error) {
	molautil_log.Debug("GetMyTrxs called")
	var trxs []*chestnutpb.Trx

	for _, blockId := range blockIds {
		block, err := nodectx.GetDbMgr().GetBlock(blockId, false, nodename)
		if err != nil {
			chain_log.Warnf(err.Error())
			continue
		}

		for _, trx := range block.Trxs {
			if trx.SenderPubkey == userSignPubkey {
				trxs = append(trxs, trx)
			}
		}
	}
	return trxs, nil
}

// get all trx from the block list
func GetAllTrxs(blocks []*chestnutpb.Block) ([]*chestnutpb.Trx, error) {
	molautil_log.Debug("GetAllTrxs called")
	var trxs []*chestnutpb.Trx
	for _, block := range blocks {
		for _, trx := range block.Trxs {
			trxs = append(trxs, trx)
		}
	}
	return trxs, nil
}

// update resend count (+1) for all trxs
func UpdateResendCount(trxs []*chestnutpb.Trx) ([]*chestnutpb.Trx, error) {
	molautil_log.Debug("UpdateResendCount called")
	for _, trx := range trxs {
		trx.ResendCount++
	}
	return trxs, nil
}
