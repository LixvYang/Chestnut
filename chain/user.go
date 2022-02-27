// Package chain provides chain for chestnut.
package chain

import (
	chestnutpb "github.com/lixvyang/chestnut/pb"
	"google.golang.org/protobuf/proto"
)

type User interface {
	Init(item *chestnutpb.GroupItem, nodename string, iface ChainMolassesIface)
	UpdAnnounce(item *chestnutpb.AnnounceItem) (string, error)
	UpdBlkList(item *chestnutpb.DenyUserItem) (string, error)
	UpdSchema(item *chestnutpb.SchemaItem) (string, error)
	UpdProducer(item *chestnutpb.ProducerItem) (string, error)
	PostToGroup(content proto.Message) (string, error)
	AddBlock(block *chestnutpb.Block) error
}