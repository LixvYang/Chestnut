// Package storage provides storage for chestnut.
package storage

type Dbmgr struct {
	GroupInfoDb ChestnutStorage
	Db	ChestnutStorage
	Auth ChestnutStorage
	DataPath string
}

