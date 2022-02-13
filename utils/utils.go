// Package utils provides the utils to the program.
package utils

import (
	"io"
	"math/rand"
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

// FileExist check if file is exist
func FileExist(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}


func DirExist(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func GetRandomStr(n int) string {
	const letters = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const lettersLength = int64(len(letters))

	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Int63()%lettersLength]
	}
	return string(b)
}

// Is DirEmpty check if dir is empty
func IsDirEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err //
}