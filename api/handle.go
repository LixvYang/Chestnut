// Package api provides API for chestnut.
package api

import (
	"context"

	"github.com/lixvyang/chestnut/appdata"
	"github.com/lixvyang/chestnut/nodectx"
	"github.com/lixvyang/chestnut/p2p"
)

type (
	Handler struct {
		Ctx context.Context
		Node *p2p.Node
		NodeCtx *nodectx.NodeCtx
		GitCommit string
		Appdb *appdata.AppDb
	}

	ErrorResponse struct {
		Error string `json:"error" validate:"required"`
	}
)