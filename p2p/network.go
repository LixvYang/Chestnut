// Package p2p provides p2p connectivity for chestnut.
package p2p

import (
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	discovery "github.com/libp2p/go-libp2p-discovery"
	"github.com/libp2p/go-libp2p-kad-dht/dual"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p-core/network"
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

