// Package p2p provides a native implementation of the p2p network protocol.
package p2p

import (
	"context"
	"fmt"

	ethkeystore "github.com/ethereum/go-ethereum/accounts/keystore"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	dsbadger2 "github.com/ipfs/go-ds-badger2"
	"github.com/libp2p/go-libp2p"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	p2pcrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/libp2p/go-libp2p-core/routing"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-kad-dht/dual"
	"github.com/libp2p/go-libp2p-peerstore/pstoreds"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	tcp "github.com/libp2p/go-tcp-transport"
	ws "github.com/libp2p/go-ws-transport"
	"github.com/lixvyang/chestnut/utils/options"
	maddr "github.com/multiformats/go-multiaddr"
)

func NewNode(ctx context.Context, nodeopt *options.NodeOptions, isBootstrap bool, ds *dsbadger2.Datastore, key *ethkeystore.Key, cmgr *connmgr.BasicConnMgr, listenAddresses []maddr.Multiaddr, jsontracerfile string) (*Node, error) {
	var ddht *dual.DHT
	var routingDiscovery *discovery.RoutingDiscovery
	var pstore peerstore.Peerstore
	var err error

	// privKey p2pcrypto.PrivKey
	ethprivkey := key.PrivateKey
	privkeybytes := ethcrypto.FromECDSA(ethprivkey)
	priv, err := p2pcrypto.UnmarshalSecp256k1PrivateKey(privkeybytes)
	if err != nil {
		return nil, err
	}

	nodenetworkname := nodeopt.NetworkName
	if nodeopt.EnableDevNetwork {
		nodenetworkname = fmt.Sprintf("%s-%s", nodeopt.NetworkName, "dev")
	}

	routingcustomprotocol := fmt.Sprintf("%s/%s", ProtocolPrefix, nodenetworkname)
	routing := libp2p.Routing(func(host host.Host) (routing.PeerRouting, error) {
		dhtOpts := dual.DHTOption(
			dht.Mode(dht.ModeServer),
			dht.Concurrency(10),
			dht.ProtocolPrefix(protocol.ID(routingcustomprotocol)),
		)
		var err error
		ddht, err = dual.New(ctx, host, dhtOpts)
		routingDiscovery = discovery.NewRoutingDiscovery(ddht)
		return ddht, err
	})
  
	networklog.Infof("Enable dht protocol prefix: %s", routingcustomprotocol)

	identity := libp2p.Identity(priv)

	libp2poptions := []libp2p.Option{
		routing,
		libp2p.ListenAddrs(listenAddresses...),
		libp2p.NATPortMap(),
		libp2p.ConnectionManager(cmgr),
		libp2p.Ping(false),
		libp2p.ChainOptions(
			libp2p.Transport(tcp.NewTCPTransport),
			libp2p.Transport(ws.New),
		),
		identity,
	}

	if ds != nil {
		pstore, err = pstoreds.NewPeerstore(ctx, ds, pstoreds.DefaultOpts())
		if err != nil {
			return nil, err
		}
		libp2poptions = append(libp2poptions, libp2p.Peerstore(pstore))
	}

	if nodeopt.EnableNat {
		libp2poptions = append(libp2poptions, libp2p.EnableNATService())
		networklog.Infof("NAT enabled")
	}

	host, err := libp2p.New(libp2poptions...)
	if err != nil {
		return nil, err
	}

	//config our own ping protocol
	pingService := &PingService{Host: host}
	host.SetStreamHandler(PingID, pingService.PingHandler)

	options := []pubsub.Option{pubsub.WithPeerExchange(true)}

	networklog.Infof("Network Name: %s", nodenetworkname)

	if isBootstrap {
		// turn off the mesh in bootstrapnode
		pubsub.GossipSubD = 0
		pubsub.GossipSubDscore = 0
		pubsub.GossipSubDlo = 0
		pubsub.GossipSubDhi = 0
		pubsub.GossipSubDout = 0
		pubsub.GossipSubDlazy = 1024
		pubsub.GossipSubGossipFactor = 0.5
	}

	var ps *pubsub.PubSub
	if jsontracerfile != "" {
		tracer, err := pubsub.NewJSONTracer(jsontracerfile)
		if err != nil {
			return nil, err
		}
		options = append(options, pubsub.WithEventTracer(tracer))
	}

	customprotocol := protocol.ID(fmt.Sprintf("%s/meshsub/1.1.0", fmt.Sprintf("%s/%s", ProtocolPrefix, nodenetworkname)))
	protos := []protocol.ID{customprotocol}
	features := func(feat pubsub.GossipSubFeature, proto protocol.ID) bool {
		if proto == customprotocol {
			return true
		}
		return false
	}

	networklog.Infof("Enable protocol: %s", customprotocol)

	options = append(options, pubsub.WithGossipSubProtocols(protos, features))
	options = append(options, pubsub.WithPeerOutboundQueueSize(128))

	ps, err = pubsub.NewGossipSub(ctx, host, options...)
	if err != nil {
		return nil, err
	}

	psping := NewPSPingService(ctx, ps, host.ID())
	psping.EnablePing()

	info := &NodeInfo{NATType: network.ReachabilityUnknown}
	newnode := &Node{NetworkName: nodenetworkname, Host: host, Pubsub: ps, Ddht: ddht, RoutingDiscovery: routingDiscovery, Info: info}

	// reconnect peers
	storedpeers := []peer.AddrInfo{}
	if ds != nil {
		for _, peer := range pstore.Peers() {
			peerinfo := pstore.PeerInfo(peer)
			storedpeers = append(storedpeers, peerinfo)
		}
	}

	if len(storedpeers) > 0 {
		//TODO: try connect every x minutes for x*y minutes?
		go func(){
			newnode.AddPeers(ctx, storedpeers)
		}()
	}

	go newnode.eventhandler(ctx)

	return newnode, nil
}