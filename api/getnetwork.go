// Package api provides API for chestnut.
package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/lixvyang/chestnut/chain"
	"github.com/lixvyang/chestnut/nodectx"
	"github.com/lixvyang/chestnut/p2p"
	"github.com/lixvyang/chestnut/utils/options"
	madrr "github.com/multiformats/go-multiaddr"
)
type groupNetworkInfo struct {
	GroupId   string    `json:"GroupId"`
	GroupName string    `json:"GroupName"`
	Peers     []peer.ID `json:"Peers"`
}

type NetworkInfo struct {
	Peerid     string                 `json:"peerid"`
	Ethaddr    string                 `json:"ethaddr"`
	NatType    string                 `json:"nat_type"`
	NatEnabled bool                   `json:"nat_enabled"`
	Addrs      []madrr.Multiaddr      `json:"addrs"`
	Groups     []*groupNetworkInfo    `json:"groups"`
	Node       map[string]interface{} `json:"node"`
}

func (h *Handler) GetNetwork(nodehost *host.Host, nodeinfo *p2p.NodeInfo, nodeopt *options.NodeOptions, ethaddr string) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		result := &NetworkInfo{}	
		node := make(map[string]interface{})
		groupnetworklist := []*groupNetworkInfo{}
		groupmgr := chain.GetGroupMgr()
		for _, group := range groupmgr.Groups {
			groupnetwork := &groupNetworkInfo{}
			groupnetwork.GroupId = group.Item.GroupId
			groupnetwork.GroupName = group.Item.GroupName
			groupnetwork.Peers = nodectx.GetNodeCtx().ListGroupPeers(group.Item.GroupId)
			groupnetworklist = append(groupnetworklist, groupnetwork)
		}
		result.Peerid = (*nodehost).ID().Pretty()
		result.Ethaddr = ethaddr
		result.NatType = nodeinfo.NATType.String()
		result.NatEnabled = nodeopt.EnableNat
		result.Addrs = (*nodehost).Addrs()
		result.Groups = groupnetworklist
		result.Node = node
		return c.JSON(http.StatusOK, result)
	}
}