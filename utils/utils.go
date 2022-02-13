// Package utils provides the utils to the program.
package utils

import (
	"os"

	logging "github.com/ipfs/go-log/v2"
)

var logger = logging.Logger("utils")

func EnsureDir(dir string) error {
	if !DirExist(dir) {
		logger.Info("Creating directory: ", dir)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			logger.Info("Error creating directory: %s ", dir)
			return err
		}
	}	
	return nil
}

func DirExist(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}
