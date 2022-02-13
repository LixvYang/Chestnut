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

	logging "github.com/ipfs/go-log/v2"
	"github.com/lixvyang/chestnut/p2p"
)

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