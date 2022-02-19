// Package storage provides storage for chestnut.
package storage

type DbMgr struct {
	GroupInfoDb ChestnutStorage
	Db	ChestnutStorage
	Auth ChestnutStorage
	DataPath string
}

