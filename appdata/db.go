// Package appdata provides storage for chestnut.
package appdata

import "github.com/lixvyang/chestnut/storage"

type AppDb struct {
	Db storage.ChestnutStorage
	seq map[string]storage.Sequence
	DataPath string
}