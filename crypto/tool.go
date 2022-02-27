// Package crypto provides crypto for chestnut.
package crypto

import "crypto/sha256"

func Hash(data []byte) []byte {
	h := sha256.New()
	h.Write(data)
	return h.Sum(nil)
}