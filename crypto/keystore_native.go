// Package crypto provides the crypto utils to the program.
package crypto

func InitKeystore(KeyStoreName, KeyStoreDir string) (int, error) {
	signkeycount := 0
	var err error
	ks, signkeycount, err = InitDirKeyStore(KeyStoreName, KeyStoreDir)
	return signkeycount, err
}