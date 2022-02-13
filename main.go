// Package main provides the entry point to the program.
package main

import (
	"context"
	"fmt"
	"os"

	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-peerstore/pstoreds"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/libp2p/go-tcp-transport"
	ws "github.com/libp2p/go-ws-transport"
	"github.com/lixvyang/chestnut/p2p"

	// pubsub "github.com/libp2p/go-libp2p-pubsub"
	// multiaddr "github.com/multiformats/go-multiaddr"
	dsbadger2 "github.com/ipfs/go-ds-badger2"
	"github.com/libp2p/go-libp2p-core/peerstore"
)

var (
	node *p2p.Node
	signalch chan os.Signal
	mainlog      = logging.Logger("main")
)

func mainRet()  {
	var pstore peerstore.Peerstore
	// var ps *pubsub.PubSub
	ctx := context.Background()

	ds, err := dsbadger2.NewDatastore("/tmp/badger", &dsbadger2.DefaultOptions)
	if err != nil {
		fmt.Errorf("Error creating datastore: %s", err)
	}
	
	libp2poptions := []libp2p.Option{
		libp2p.NATPortMap(),
		libp2p.ChainOptions(
			libp2p.Transport(tcp.NewTCPTransport),
			libp2p.Transport(ws.New),
		),
	}

	if ds != nil {
		pstore, err = pstoreds.NewPeerstore(ctx, ds, pstoreds.DefaultOpts())
		if err != nil {
			fmt.Errorf("Error creating peerstore: %s", err)
		}
		libp2poptions = append(libp2poptions, libp2p.Peerstore(pstore))
	}
	node, err := libp2p.New(
		libp2poptions...,
	)
	if err != nil {
		panic(err)
	}
	//config our own ping protocol
	pingService := &ping.PingService{Host:node}
	node.SetStreamHandler(ping.ID, pingService.PingHandler)
	fmt.Println(node)
	// options := []pubsub.Option{pubsub.WithPeerExchange(true)}
	// ps, err = pubsub.NewGossipSub(ctx, node, options...)
	// if err != nil {
	// 	fmt.Errorf("Error creating pubsub: %s", err)
	// }
	
	if err := node.Close(); err != nil {
		panic(err)
	}
}

func main()  {
	

	// mainRet()
}