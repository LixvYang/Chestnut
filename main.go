// Package main provides the entry point to the program.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	ethkeystore "github.com/ethereum/go-ethereum/accounts/keystore"
	dsbadger2 "github.com/ipfs/go-ds-badger2"
	logging "github.com/ipfs/go-log/v2"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	discovery "github.com/libp2p/go-libp2p-discovery"
	"github.com/lixvyang/chestnut/api"
	"github.com/lixvyang/chestnut/appdata"
	"github.com/lixvyang/chestnut/chain"
	localcrypto "github.com/lixvyang/chestnut/crypto"
	"github.com/lixvyang/chestnut/nodectx"
	"github.com/lixvyang/chestnut/p2p"
	appapi "github.com/lixvyang/chestnut/pkg/app/api"
	"github.com/lixvyang/chestnut/storage"
	"github.com/lixvyang/chestnut/utils/cli"
	"github.com/lixvyang/chestnut/utils/options"
)

const DEFAULT_KEY_NAME = "default"

var (
	ReleaseVersion string
	GitCommit      string
	node *p2p.Node
	signalch chan os.Signal
	mainlog      = logging.Logger("main")
)

// reutrn EBUSY if LOCK is exist
func checkLockError(err error) {
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "Another process is using this Badger database.") {
			mainlog.Errorf(errStr)
			os.Exit(16)
		}
	}
}

func createDb(path string) (*storage.DbMgr, error) {
	var err error
	groupDb := storage.CSBadger{}
	dataDb := storage.CSBadger{}
	err = groupDb.Init(path)
	if err != nil {
		return nil, err
	}

	err = dataDb.Init(path)
	if err != nil {
		return nil, err
	}

	manager := storage.DbMgr{&groupDb, &dataDb, nil, path}
	return &manager, nil
}

func createAppDb(path string) (*appdata.AppDb, error) {
	var err error
	db := storage.CSBadger{}
	err = db.Init(path + "_appdb")
	if err != nil {
		return nil, err
	}

	app := appdata.NewAppDb()
	app.Db = &db
	app.DataPath = path
	return app, nil
}

// mainRet is the main function for the program. It is called from main.
func mainRet(config cli.Config) int {
	signalch = make(chan os.Signal, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	peername := config.PeerName
	if config.IsBootstrap {
		peername = "bootstrap"
	}

	nodeoptions, err := options.InitNodeOptions(config.ConfigDir, peername)
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

	ds, err := dsbadger2.NewDatastore(path.Join(config.ConfigDir, fmt.Sprintf("%s-%s", peername, "peerstore")), &dsbadger2.DefaultOptions)
	checkLockError(err)
	if err != nil {
		cancel()
		mainlog.Fatalf(err.Error())
	}

	if config.IsBootstrap {
		// bootstrop node connections: low watermarks: 1000 high watermarks 50000, grace 30s
		connmanager, _ := connmgr.NewConnManager(1000, 50000, connmgr.WithGracePeriod(30 * time.Second), connmgr.WithEmergencyTrim(true))
		node, err := p2p.NewNode(ctx, nodeoptions, config.IsBootstrap, ds, defaultkey, connmanager, config.ListenAddresses, config.JsonTracer)
		if err != nil {
			mainlog.Fatalf(err.Error())
		}
		datapath := config.DataDir + "/" + config.PeerName
		
		dbManager, err := createDb(datapath)
		if err != nil {
			mainlog.Fatalf(err.Error())
		}

		nodectx.InitCtx(ctx, "", node, dbManager, "pubsub", GitCommit)
		nodectx.GetNodeCtx().Keystore = ksi
		nodectx.GetNodeCtx().PublickKey = keys.PubKey
		nodectx.GetNodeCtx().PeerId = peerid

		mainlog.Infof("Host created, ID:<%s>, Address:<%s>", node.Host.ID(), node.Host.Addrs())
		h := &api.Handler{Node: node, NodeCtx: nodectx.GetNodeCtx(), GitCommit: GitCommit}
		go api.StartAPIServer(config, signalch, h, nil, node, nodeoptions, ks, ethaddr, true)
	} else {
		//normal node connections: low watermarks: 10  hi watermarks 200, grace 60s
		connmanager, _ := connmgr.NewConnManager(10, 200, connmgr.WithGracePeriod(60 * time.Second), connmgr.WithEmergencyTrim(true))
		node, err = p2p.NewNode(ctx, nodeoptions, config.IsBootstrap, ds, defaultkey, connmanager, config.ListenAddresses, config.JsonTracer)
		_ = node.Bootstrap(ctx, config)

		for _, addr := range node.Host.Addrs() {
			p2paddr := fmt.Sprintf("%s/p2p/%s", addr.String(), node.Host.ID())
			mainlog.Infof("Peer ID:<%s>, Peer Address:<%s>", node.Host.ID(), p2paddr)
		}

		//Discovery and Advertise had been replaced by PeerExchange
		mainlog.Infof("Announcing ourselves...")
		discovery.Advertise(ctx, node.RoutingDiscovery, config.RendezvousString)
		mainlog.Infof("Successfully announced!")
		peerok := make(chan struct{})

		go node.ConnectPeers(ctx, peerok, nodeoptions.MaxPeers, config)
		datapath := config.DataDir + "/" + config.PeerName
		dbManager, err := createDb(datapath)
		if err != nil {
			mainlog.Fatalf(err.Error())
		}
		dbManager.TryMigration(0)
		nodectx.InitCtx(ctx, "default", node, dbManager, "pubsub", GitCommit)
		nodectx.GetNodeCtx().Keystore = ksi
		nodectx.GetNodeCtx().PublickKey = keys.PubKey
		nodectx.GetNodeCtx().PeerId = peerid
		groupmgr := chain.InitGroupMgr(nodectx.GetDbMgr())

		err = groupmgr.SyncAllGroup()
		if err != nil {
			mainlog.Fatalf(err.Error())
		}

		appdb, err := createAppDb(datapath)
		if err != nil {
			mainlog.Fatalf(err.Error())
		}
		checkLockError(err)

		// run local http api service
		h := &api.Handler{
			Node: node,
			NodeCtx: nodectx.GetNodeCtx(),
			Ctx: ctx,
			GitCommit: GitCommit,
			Appdb: appdb,
		}

		apiaddress := "http://%s/api/v1"
		if config.APIListenAddresses[:1] == ":" {
			apiaddress = fmt.Sprintf(apiaddress, "localhost"+config.APIListenAddresses)
		} else {
			apiaddress = fmt.Sprintf(apiaddress, config.APIListenAddresses)
		}

		appsync := appdata.NewAppSyncAgent(apiaddress, "default", appdb, dbManager)
		appsync.Start(10)
		apph := &appapi.Handler{
			Appdb: appdb,
			Chaindb: dbManager,
			GitCommit: GitCommit,
			Apiroot: apiaddress,
			ConfigDir: config.ConfigDir,
			PeerName: config.PeerName,
			NodeName: nodectx.GetNodeCtx().Name,
		}
		go api.StartAPIServer(config, signalch, h, apph, node, nodeoptions, ks, ethaddr, false)
	}

	//attach signal
	signal.Notify(signalch, os.Interrupt, os.Kill, syscall.SIGTERM)
	signalType := <- signalch
	signal.Stop(signalch)

	if !config.IsBootstrap {
		groupmgr := chain.GetGroupMgr()
		groupmgr.Release()
	}
	// clean up before exit
	mainlog.Infof("On Signal <%s>", signalType)
	mainlog.Infof("Exit command received. Exiting...")
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