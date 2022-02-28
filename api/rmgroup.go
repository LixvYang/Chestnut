// Package api provides API for chestnut.
package api

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	p2pcrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/lixvyang/chestnut/chain"
	localcrypto "github.com/lixvyang/chestnut/crypto"
	"github.com/lixvyang/chestnut/nodectx"
)


type RmGroupParam struct {
	GroupId string `from:"group_id" json:"group_id" validate:"required"`
}

type RmGroupResult struct {
	GroupId     string `json:"group_id"`
	Signature   string `json:"signature"`
	OwnerPubkey string `json:"owner_pubkey"`
}


func (h *Handler) RmGroup(c echo.Context) (err error) {
	output := make(map[string]string)
	validate := validator.New()
	params := new(RmGroupParam)
	if err = c.Bind(params); err != nil {
		output[ERROR_INFO] = err.Error()
		return c.JSON(http.StatusBadRequest, output)
	}

	if err = validate.Struct(params); err != nil {
		output[ERROR_INFO] = err.Error()
		return c.JSON(http.StatusBadRequest,output)
	}

	shouldRemove := false
	groupmgr := chain.GetGroupMgr()
	if group, ok := groupmgr.Groups[params.GroupId]; ok {
		err := group.DelGrp()

		if err != nil {
			output[ERROR_INFO] = err.Error()
			return c.JSON(http.StatusBadRequest,output)
		}

		shouldRemove = true
	} else {
		output[ERROR_INFO] = fmt.Sprintf("Group %s not exist", params.GroupId)
		return c.JSON(http.StatusBadRequest, output)
	}

	if shouldRemove {
		delete(groupmgr.Groups, params.GroupId)
	}

	var groupSignPubkey []byte
	ks := nodectx.GetNodeCtx().Keystore
	dirks, ok := ks.(*localcrypto.DirKeyStore)
	if !ok {
		//use "default" key for all groups
		//TODO: user can create new sign keys for each groups
		hexkey, err := dirks.GetEncodedPubkey(params.GroupId, localcrypto.Sign)
		pubkeybytes, err := hex.DecodeString(hexkey)
		p2ppubkey, err := p2pcrypto.UnmarshalSecp256k1PublicKey(pubkeybytes)
		groupSignPubkey, err = p2pcrypto.MarshalPublicKey(p2ppubkey)
		if err != nil {
			output[ERROR_INFO] = "group key can't be decoded, err:" + err.Error()
			return c.JSON(http.StatusBadRequest, output)
		}
	} else {
		output[ERROR_INFO] = fmt.Sprintf("unknown keystore type  %v:", ks)
		return c.JSON(http.StatusBadRequest, output)
	}

	var buffer bytes.Buffer
	buffer.Write(groupSignPubkey)
	buffer.Write([]byte(params.GroupId))
	hash := chain.Hash(buffer.Bytes())
	signature, err := ks.SignByKeyName(params.GroupId, hash)
	encodeSign := hex.EncodeToString(signature)
	result := &RmGroupResult{
		GroupId:     params.GroupId,
		Signature:   encodeSign,
		OwnerPubkey: p2pcrypto.ConfigEncodeKey(groupSignPubkey),
	}
	return c.JSON(http.StatusOK, result)
}
