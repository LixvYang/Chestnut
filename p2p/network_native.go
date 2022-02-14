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
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/libp2p/go-libp2p-core/routing"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-kad-dht/dual"
	tcp "github.com/libp2p/go-tcp-transport"
	"github.com/lixvyang/chestnut/utils/options"
	maddr "github.com/multiformats/go-multiaddr"
	ws "github.com/libp2p/go-ws-transport"
	"github.com/libp2p/go-libp2p-peerstore/pstoreds"
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
}