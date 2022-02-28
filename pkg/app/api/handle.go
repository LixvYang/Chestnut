package api

import (
	"context"

	"github.com/lixvyang/chestnut/appdata"
	"github.com/lixvyang/chestnut/storage"
)

type Handler struct {
	Ctx       context.Context
	Appdb     *appdata.AppDb
	Chaindb   *storage.DbMgr
	Apiroot   string
	GitCommit string
	ConfigDir string
	PeerName  string
	NodeName  string
}
