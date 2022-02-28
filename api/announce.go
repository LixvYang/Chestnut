// Package api provides API for chestnut.
package api

import (
	"bytes"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/lixvyang/chestnut/chain"
	localcrypto "github.com/lixvyang/chestnut/crypto"
	"github.com/lixvyang/chestnut/nodectx"
	chestnutpb "github.com/lixvyang/chestnut/pb"
)

type AnnounceResult struct {
	GroupId                string `json:"group_id"`
	AnnouncedSignPubkey    string `json:"sign_pubkey"`
	AnnouncedEncryptPubkey string `json:"encrypt_pubkey"`
	Type                   string `json:"type"`
	Action                 string `json:"action"`
	Sign                   string `json:"sign"`
	TrxId                  string `json:"trx_id"`
}

type AnnounceParam struct {
	GroupId string `from:"group_id"    json:"group_id"    validate:"required"`
	Action  string `from:"action"      json:"action"      validate:"required,oneof=add remove"`
	Type    string `from:"type"        json:"type"        validate:"required,oneof=user producer"`
	Memo    string `from:"memo"        json:"memo"        validate:"required"`
}

func (h *Handler) Announce(c echo.Context) (err error) {
	output := make(map[string]string)
	validate := validator.New()
	params := new(AnnounceParam)

	if err = c.Bind(params); err != nil {
		output[ERROR_INFO] = err.Error()
		return c.JSON(http.StatusBadRequest, output)
	}

	if err = validate.Struct(params); err != nil {
		output[ERROR_INFO] = err.Error()
		return c.JSON(http.StatusBadRequest, output)
	}

	 item := &chestnutpb.AnnounceItem{}
	 item.GroupId = params.GroupId

	 groupmgr := chain.GetGroupMgr()
	 if group, ok := groupmgr.Groups[params.GroupId]; !ok {
			output[ERROR_INFO] = "Can not find group"
			return c.JSON(http.StatusBadRequest, output)
	 } else {
			if params.Type == "user" {
				item.Type = chestnutpb.AnnounceType_AS_USER
			} else if params.Type == "producer" {
				item.Type = chestnutpb.AnnounceType_AS_PRODUCER
			} else {
				output[ERROR_INFO] = "Unknown type"
				return c.JSON(http.StatusBadRequest, output)
			}

			if params.Action == "add" {
				item.Action = chestnutpb.ActionType_ADD
			} else if params.Action == "remove" {
				item.Action = chestnutpb.ActionType_REMOVE
			} else {
				output[ERROR_INFO] = "Unknown action"
				return c.JSON(http.StatusBadRequest, output)
			}

			item.SignPubkey = group.Item.UserSignPubkey

			if item.Type == chestnutpb.AnnounceType_AS_USER {
				item.EncryptPubkey, err = nodectx.GetNodeCtx().Keystore.GetEncodedPubkey(params.GroupId, localcrypto.Encrypt)
			}

			if err != nil {
				output[ERROR_INFO] = err.Error()
				return c.JSON(http.StatusBadRequest, output)
			}

			item.OwnerPubkey = ""
			item.OwnerSignature = ""
			item.Result = chestnutpb.ApproveType_ANNOUNCED

			var buffer bytes.Buffer
			buffer.Write([]byte(item.GroupId))
			buffer.Write([]byte(item.SignPubkey))
			buffer.Write([]byte(item.EncryptPubkey))
			buffer.Write([]byte(item.Type.String()))
			hash := chain.Hash(buffer.Bytes())
			signature, err := nodectx.GetNodeCtx().Keystore.SignByKeyName(item.GroupId, hash)
			if err != nil {
				output[ERROR_INFO] = err.Error()
				return c.JSON(http.StatusBadRequest, output)
			}

			item.AnnouncerSignature = hex.EncodeToString(signature)
			item.TimeStamp = time.Now().UnixNano()

			trxId, err := group.UpdAnnounce(item)

			if err != nil {
				output[ERROR_INFO] = err.Error()
				return c.JSON(http.StatusBadRequest, output)
			}

			var announceResult *AnnounceResult
			announceResult = &AnnounceResult{GroupId: item.GroupId, AnnouncedSignPubkey: item.SignPubkey, AnnouncedEncryptPubkey: item.EncryptPubkey, Type: item.Type.String(), Action: item.Action.String(), Sign: hex.EncodeToString(signature), TrxId: trxId}

			return c.JSON(http.StatusOK, announceResult)
		}
}