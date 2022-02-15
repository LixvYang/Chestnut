// Package p2p provides p2p connectivity for chestnut.
package p2p

import (
	"context"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/event"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	discovery "github.com/libp2p/go-libp2p-discovery"
	"github.com/libp2p/go-libp2p-kad-dht/dual"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

const ProtocolPrefix = "/chestnut"

var networklog = logging.Logger("network")

type NodeInfo struct {
	NATType network.Reachability
}

type Node struct {
	PeerID peer.ID
	Host host.Host
	NetworkName string
	Pubsub *pubsub.PubSub
	Ddht *dual.DHT
	Info *NodeInfo
	RoutingDiscovery *discovery.RoutingDiscovery
}

func (node *Node) eventhandler(ctx context.Context) {
	evbus := node.Host.EventBus()
	subReachability, err := evbus.Subscribe(new(event.EvtLocalReachabilityChanged))
	if err != nil {
		networklog.Errorf("event subcribe err: %s", err)
	}
	defer subReachability.Close()
	for {
		select {
		case ev := <- subReachability.Out():
			evt, ok := ev.(event.EvtLocalReachabilityChanged)
			if !ok {
				return
			}
			networklog.Infof("Reachability chante: %s", evt.Reachability.String())
			node.Info.NATType = evt.Reachability
		case <-ctx.Done():
			return 
		}
	}
}

func (node *Node) AddPeers(ctx context.Context, peers []peer.AddrInfo) int {
	connectedCount := 0
	for _, peer := range peers {
		if peer.ID == node.Host.ID() {
			continue
		}
		pctx, cancel := context.WithTimeout(ctx, time.Second*10)
		defer cancel()
		err := node.Host.Connect(pctx, peer)
		if err != nil {
			networklog.Warnf("Connect to peer failure: %s", peer)
			cancel()
			continue 
		} else {
			connectedCount++
			networklog.Info("Connect: %s", peer)
		}
	}
	return connectedCount
}
