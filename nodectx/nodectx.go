// Package nodectx provides context for node.
package nodectx

import (
	"context"

	logging "github.com/ipfs/go-log/v2"
	p2pcrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	localcrypto "github.com/lixvyang/chestnut/crypto"
	"github.com/lixvyang/chestnut/p2p"
	"github.com/lixvyang/chestnut/storage"
)

var chainctx_log = logging.Logger("chainctx")

type NodeStatus int8

const (
	USER_CHANNEL_PREFIX     = "user_channel_"
	PRODUCER_CHANNEL_PREFIX = "prod_channel_"

	NODE_ONLINE  = 0
	NODE_OFFLINE = 1
)

type NodeCtx struct {
	Node *p2p.Node
	PeerId peer.ID
	Keystore localcrypto.Keystore
	PublickKey p2pcrypto.PubKey
	Name string
	Ctx context.Context
	Version string
	Status NodeStatus
	Economic int
}

var (
	nodeCtx *NodeCtx
	dbMgr *storage.DbMgr
)

func GetNodeCtx() *NodeCtx {
	return nodeCtx
}

func GetDbMgr() *storage.DbMgr {
	return dbMgr
}

func InitCtx(ctx context.Context, name string, node *p2p.Node, db *storage.DbMgr, channeltype string, gitcommit string)  {
	nodeCtx := &NodeCtx{}
	nodeCtx.Ctx = ctx
	nodeCtx.Node = node
	dbMgr = db

	nodeCtx.Status = NODE_OFFLINE
	nodeCtx.Name = name
	nodeCtx.Version = "1.0.0"
}

func (nodeCtx *NodeCtx) PeersProtocol() *map[string][]string {
	return nodeCtx.Node.PeersProtocol()
}

func (nodeCtx *NodeCtx) ProtocolPrefix() string {
	return p2p.ProtocolPrefix
}

func (nodeCtx *NodeCtx) UpdateOnlineStatus(status NodeStatus)  {
	nodeCtx.Status = status
}

func (nodeCtx *NodeCtx) GetNodePubKey() (string, error) {
	var pubkey string
	pubkeybytes, err := p2pcrypto.MarshalPublicKey(nodeCtx.PublickKey)
	if err != nil {
		return pubkey, err
	}
	pubkey = p2pcrypto.ConfigEncodeKey(pubkeybytes)
	return pubkey, nil
}

func (nodeCtx *NodeCtx) ListGroupPeers(groupid string) []peer.ID {
	userChannelId := USER_CHANNEL_PREFIX + groupid
	return nodeCtx.Node.Pubsub.ListPeers(userChannelId)
}

func (nodeCtx *NodeCtx) AddPeers(peers []peer.AddrInfo) int {
	return nodeCtx.Node.AddPeers(nodeCtx.Ctx, peers)
}

