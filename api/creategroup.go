// Package api provides API for chestnut.
package api

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	guuid "github.com/google/uuid"
	"github.com/labstack/echo/v4"
	p2pcrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/lixvyang/chestnut/chain"
	localcrypto "github.com/lixvyang/chestnut/crypto"
	"github.com/lixvyang/chestnut/nodectx"
	chestnutpb "github.com/lixvyang/chestnut/pb"
	"github.com/lixvyang/chestnut/utils/options"
)

type CreateGroupParam struct {
	GroupName      string `from:"group_name" json:"group_name" validate:"required,max=20,min=5"`
	ConsensusType  string `from:"consensus_type"  json:"consensus_type"  validate:"required,oneof=pos poa"`
	EncryptionType string `from:"encryption_type" json:"encryption_type" validate:"required,oneof=public private"`
	AppKey         string `from:"app_key"         json:"app_key"         validate:"required,max=20,min=5"`
}

type CreateGroupResult struct {
	GenesisBlock       *chestnutpb.Block `json:"genesis_block"`
	GroupId            string          `json:"group_id"`
	GroupName          string          `json:"group_name"`
	OwnerPubkey        string          `json:"owner_pubkey"`
	OwnerEncryptPubkey string          `json:"owner_encryptpubkey"`
	ConsensusType      string          `json:"consensus_type"`
	EncryptionType     string          `json:"encryption_type"`
	CipherKey          string          `json:"cipher_key"`
	AppKey             string          `json:"app_key"`
	Signature          string          `json:"signature"`
}


func (h *Handler) CreateGroup(c echo.Context) (err error)  {
	output := make(map[string]string)

	validate := validator.New()
	params := new(CreateGroupParam)
	if err = c.Bind(params); err != nil {
		output[ERROR_INFO] = err.Error()
		return c.JSON(http.StatusBadRequest, output)
	}

	if err = validate.Struct(params); err != nil {
		output[ERROR_INFO] = err.Error()
		return c.JSON(http.StatusBadRequest, output)
	}

	if params.ConsensusType != "poa" {
		output := "Other types of groups are not supported yet"
		return c.JSON(http.StatusBadRequest, output)
	}

	groupid := guuid.New()

	nodeoptions := options.GetNodeOptions()

	var groupSignPubkey []byte
	var p2ppubkey p2pcrypto.PubKey
	ks := nodectx.GetNodeCtx().Keystore
	dirks, ok := ks.(*localcrypto.DirKeyStore)
	if ok {
		hexkey, err := dirks.GetEncodedPubkey(groupid.String(), localcrypto.Sign)
		if err != nil && strings.HasPrefix(err.Error(), "key not exist "){
			newsignaddr, err := dirks.NewKeyWithDefaultPassword(groupid.String(),localcrypto.Sign)
			if err != nil && newsignaddr != "" {
				err = nodeoptions.SetSignKeyMap(groupid.String(), newsignaddr)
				if err != nil {
					output[ERROR_INFO] = fmt.Sprintf("save key map %s err : %s",newsignaddr,err.Error())
					return c.JSON(http.StatusBadRequest, output)
				}
			}
			hexkey, err = dirks.GetEncodedPubkey(groupid.String(), localcrypto.Sign)
		} else {
			output[ERROR_INFO] = "Create new group key err:" + err.Error()
			return c.JSON(http.StatusBadRequest, output)
		}
		pubkeybytes, err := hex.DecodeString(hexkey)
		p2ppubkey, err := p2pcrypto.UnmarshalSecp256k1PublicKey(pubkeybytes)
		groupSignPubkey, err = p2pcrypto.MarshalPublicKey(p2ppubkey)
		if err != nil {
			output[ERROR_INFO] = "group key can't be decoded err:" + err.Error()
			return c.JSON(http.StatusBadRequest, output)
		}
	} else {
		output[ERROR_INFO] = fmt.Sprintf("unknown keystore type  %v:", ks)
		return c.JSON(http.StatusBadRequest, output)
	}

	genesisBlock, err := chain.CreateGeneisBlock(groupid.String(), p2ppubkey)
	if err != nil {
		output[ERROR_INFO] = "Create genesis block err:" + err.Error()
		return c.JSON(http.StatusBadRequest, output)
	}

	genesisBlockBytes, err := json.Marshal(genesisBlock)
	if err != nil {
		output[ERROR_INFO] = err.Error()
		return c.JSON(http.StatusBadRequest, output)
	}

	cipherKey, err := localcrypto.CreateAesKey()
	
	groupEncryptPubkey, err := dirks.GetEncodedPubkey(groupid.String(), localcrypto.Encrypt)
	if err != nil {
		if strings.HasPrefix(err.Error(), "key not exist "){
			groupEncryptPubkey, err = dirks.NewKeyWithDefaultPassword(groupid.String(), localcrypto.Encrypt)
			if err != nil {
				output[ERROR_INFO] = "Create key pair failed with msg:" + err.Error()
				return c.JSON(http.StatusBadRequest, output)
			}
		} else {
			output[ERROR_INFO] = "Create key pair failed with msg:" + err.Error()
			return c.JSON(http.StatusBadRequest, output)
		}
	}

	var item *chestnutpb.GroupItem
	item = &chestnutpb.GroupItem{}
	item.GroupId = groupid.String()
	item.GroupName = params.GroupName
	item.OwnerPubKey = p2pcrypto.ConfigEncodeKey(groupSignPubkey)
	item.UserSignPubkey = item.OwnerPubKey
	item.UserEncryptPubkey = groupEncryptPubkey
	item.ConsenseType = chestnutpb.GroupConsenseType_POA

	if params.EncryptionType == "public" {
		item.EncryptType = chestnutpb.GroupEncryptType_PUBLIC
	} else {
		item.EncryptType = chestnutpb.GroupEncryptType_PRIVATE
	}

	item.CipherKey = hex.EncodeToString(cipherKey)
	item.AppKey = params.AppKey
	item.HighestHeight = 0
	item.HighestBlockId = genesisBlock.BlockId
	item.LastUpdate = time.Now().UnixNano()
	item.GenesisBlock = genesisBlock

	var group *chain.Group
	group = &chain.Group{}

	err = group.CreateGrp(item)
	if err != nil {
		output[ERROR_INFO] = err.Error()
		return c.JSON(http.StatusBadRequest, output)
	}

	groupmgr := chain.GetGroupMgr()
	groupmgr.Groups[group.Item.GroupId] = group
	
	//create result
	var buffer bytes.Buffer
	buffer.Write(genesisBlockBytes)
	buffer.Write([]byte(groupid.String()))
	buffer.Write([]byte(params.GroupName))
	buffer.Write(groupSignPubkey) //group owner pubkey
	buffer.Write([]byte(params.ConsensusType))
	buffer.Write([]byte(params.EncryptionType))
	buffer.Write([]byte(params.AppKey))
	buffer.Write(cipherKey)

	hash := localcrypto.Hash(buffer.Bytes())
	signature, err := ks.SignByKeyName(groupid.String(), hash)
	encodedSign := hex.EncodeToString(signature)
	encodedCipherKey := hex.EncodeToString(cipherKey)

	CreateGroupParam  := &CreateGroupResult{
		GenesisBlock: genesisBlock,
		GroupId: groupid.String(),
		GroupName: params.GroupName,
		OwnerPubkey: item.OwnerPubKey,
		OwnerEncryptPubkey: item.UserEncryptPubkey,
		ConsensusType: params.ConsensusType,
		EncryptionType: params.EncryptionType,
		CipherKey: encodedCipherKey,
		AppKey: params.AppKey,
		Signature: encodedSign,
	}
	return c.JSON(http.StatusOK, CreateGroupParam)
}