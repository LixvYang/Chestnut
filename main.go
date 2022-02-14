// Package main provides the entry point to the program.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/lixvyang/chestnut/utils/cli"
	"github.com/lixvyang/chestnut/utils/options"
	localcrypto "github.com/lixvyang/chestnut/crypto"
	ethkeystore "github.com/ethereum/go-ethereum/accounts/keystore"

	logging "github.com/ipfs/go-log/v2"
	"github.com/lixvyang/chestnut/p2p"
)

const DEFAULT_KEY_NAME = "default"

var (
	node *p2p.Node
	signalch chan os.Signal
	mainlog      = logging.Logger("main")
)

// mainRet is the main function for the program. It is called from main.
func mainRet(config cli.Config) int {
	signalch = make(chan os.Signal, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	peername := config.PeerName
	if config.IsBootstrap {
		peername = "bootstrap"
	}

	nodeoptions, err := options.GetNodeOptions(config.ConfigDir, peername)
	if err != nil {
		cancel()
		mainlog.Fatalf(err.Error())
	}

	signkeycount, err := localcrypto.InitKeystore(config.KeyStoreName, config.KeyStoreDir)
	ksi := localcrypto.GetKeystore()
	if err != nil {
		cancel()
		mainlog.Fatalf(err.Error())
	}

	ks, ok := ksi.(*localcrypto.DirKeyStore)
	if !ok {
		cancel()
		mainlog.Fatalf(err.Error())
	}

	password := os.Getenv("CHESTNUT_PASSWORD")
	if signkeycount > 0 {
		if password == "" {
			password, err = localcrypto.PassphrasePromptForUnlock()
		}
		err = ks.Unlock(nodeoptions.SignKeyMap, password)
		if err != nil {
			mainlog.Fatalf(err.Error())
			cancel()
			return 0
		}
	} else {
		if password == "" {
			password, err = localcrypto.PassphrasePromptForEncryption()
			if err != nil {
				mainlog.Fatalf(err.Error())
				cancel()
				return 0
			}
			fmt.Println("Please keeping your password safe, We can't recover or reset your password.")
			fmt.Println("Your password:", password)
			fmt.Println("After saving the password, press any key to continue.")
			os.Stdin.Read(make([]byte, 1))
		}

		signkeyhexstr, err := localcrypto.LoadEncodeKeyFrom(config.ConfigDir, peername, "txt")
		if err != nil {
			cancel()
			mainlog.Fatalf(err.Error())
		}

		var addr string
		if signkeyhexstr != "" {
			addr, err = ks.Import(DEFAULT_KEY_NAME, signkeyhexstr, localcrypto.Sign, password)
		} else {
			addr, err = ks.NewKey(DEFAULT_KEY_NAME, localcrypto.Sign, password)
			if err != nil {
				cancel()
				mainlog.Errorf(err.Error())
				return 0
			}
		}

		if addr == "" {
			cancel()
			mainlog.Errorf("Load or create new signkey failed")
			return 0
		}

		err = nodeoptions.SetSignKeyMap(DEFAULT_KEY_NAME, addr)
		if err != nil {
			cancel()
			mainlog.Errorf(err.Error())
			return 0
		}
		err = ks.Unlock(nodeoptions.SignKeyMap, password)
		if err != nil {
			mainlog.Fatalf(err.Error())
			cancel()
			return 0
		}

		fmt.Printf("load signkey: %d press any key to continue...\n", signkeycount)
	}
	_, err = ks.GetKeyFromUnlocked(localcrypto.Sign.NameString(DEFAULT_KEY_NAME))
	signkeycount = ks.UnlockedKeyCount(localcrypto.Sign)
	if signkeycount == 0 {
		mainlog.Fatalf("load signkey error, exit... %s", err)
		cancel()
		return 0
	}

	// Load default sign keys
	key, err := ks.GetKeyFromUnlocked(localcrypto.Sign.NameString(DEFAULT_KEY_NAME))

	defaultkey, ok := key.(*ethkeystore.Key)
	if !ok {
		fmt.Println("load default key error, exit...")
		cancel()
		mainlog.Errorf(err.Error())
		return 0
	}

	keys, err := localcrypto.SignKeytoPeerKeys(defaultkey)
	if err != nil {
		cancel()
		mainlog.Fatalf(err.Error())
		return 0
	}

	peerid, ethaddr, err := ks.GetPeerInfo(DEFAULT_KEY_NAME)
	if err != nil {
		cancel()
		mainlog.Fatalf(err.Error())
	}

	mainlog.Infof("eth address: <%s>", ethaddr)

	


	return 0
}

func main()  {
	help := flag.Bool("h", false, "Display help")

	config, err := cli.ParseFlags()
	if err != nil {
		panic(err)
	}

	if config.IsDebug {
		logging.SetLogLevel("main", "debug")
		logging.SetLogLevel("crypto", "debug")
		logging.SetLogLevel("network", "debug")
		logging.SetLogLevel("pubsub", "debug")
		logging.SetLogLevel("autonat", "debug")
		logging.SetLogLevel("chain", "debug")
		logging.SetLogLevel("dbmgr", "debug")
		logging.SetLogLevel("chainctx", "debug")
		logging.SetLogLevel("group", "debug")
		logging.SetLogLevel("syncer", "debug")
		logging.SetLogLevel("producer", "debug")
		logging.SetLogLevel("user", "debug")
		logging.SetLogLevel("groupmgr", "debug")
		logging.SetLogLevel("trxmgr", "debug")
	}

	if *help {
		fmt.Println("Output a help ")
		fmt.Println()
		fmt.Println("Useage...")
		flag.PrintDefaults()
		return
	}


	os.Exit(mainRet(config))
}