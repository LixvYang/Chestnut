// Package api provides API for chestnut.
package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/lixvyang/chestnut/handlers"
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
	
	params := new(handlers.CreateGroupParam)
	if err = c.Bind(params); err != nil {
		output[ERROR_INFO] = err.Error()
		return c.JSON(http.StatusBadRequest, output)
	}

	res, err := handlers.CreateGroup(params, options.GetNodeOptions(), h.Appdb)
	if err != nil {
		output[ERROR_INFO] = err.Error()
		return c.JSON(http.StatusBadRequest, output)
	}

	return c.JSON(http.StatusOK, res)
}