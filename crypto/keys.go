// Package crypto provides the crypto utils to the program.
package crypto

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"filippo.io/age"
	logging "github.com/ipfs/go-log/v2"
	p2pcrypto "github.com/libp2p/go-libp2p-core/crypto"
	ethkeystore "github.com/ethereum/go-ethereum/accounts/keystore"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

var cryptolog = logging.Logger("crypto")

type Key struct {
	PrivKey p2pcrypto.PrivKey
	PubKey  p2pcrypto.PubKey
	EthAddr string
	groupKeys map[string]*age.X25519Identity
}

func LoadEncodeKeyFrom(dir, keyname, filetype string) (string, error) {
	keyfilepath := filepath.FromSlash(fmt.Sprintf("%s/%s_keys.%s", dir, keyname, filetype))
	if filetype == "txt" {
		f, err := os.Open(keyfilepath)
		if err != nil {
			if os.IsNotExist(err) {
				return "", nil
			}
			return "", err
		}
		defer f.Close()
		buf, err := ioutil.ReadAll(f)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(buf)), nil
	} else {
		return "", fmt.Errorf("unsupported filetype %s", filetype)
	}
}

func SignKeytoPeerKeys(key *ethkeystore.Key) (*Key, error) {
	println("开始进入公钥密钥")
	ethprivkey := key.PrivateKey
	privkeybytes := ethcrypto.FromECDSA(ethprivkey)
	pubkeybytes := ethcrypto.FromECDSAPub(&ethprivkey.PublicKey)

	priv, err := p2pcrypto.UnmarshalECDSAPrivateKey(privkeybytes)
	pub, err := p2pcrypto.UnmarshalECDSAPublicKey(pubkeybytes)
	if err != nil {
		return nil, err
	}
	println("公钥 密钥 完毕")
	address := ethcrypto.PubkeyToAddress(ethprivkey.PublicKey).Hex()
	return &Key{PrivKey: priv, PubKey: pub, EthAddr: address}, nil
}