// Package api provides API for chestnut.
package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/lixvyang/chestnut/nodectx"
	"github.com/lixvyang/chestnut/p2p"
	p2pcrypto "github.com/libp2p/go-libp2p-core/crypto"
)



type NodeInfo struct {
	Node_publickey string `json:"node_publickey"`
	Node_status    string `json:"node_status"`
	Node_version   string `json:"node_version"`
	User_id        string `json:"user_id"`
}

func updateNodeStatus(nodenetworkname string) {
	peersprotocol := nodectx.GetNodeCtx().PeersProtocol()
	networknamewithprefix := fmt.Sprintf("%s/%s", p2p.ProtocolPrefix, nodenetworkname)
	for protocol, peerlist := range *peersprotocol {
		if strings.HasPrefix(protocol, networknamewithprefix) {
			if len(peerlist) > 0 {
				nodectx.GetNodeCtx().UpdateOnlineStatus(nodectx.NODE_ONLINE)
				return
			}
		}
	}
	if nodectx.GetNodeCtx().Status != nodectx.NODE_OFFLINE {
		nodectx.GetNodeCtx().UpdateOnlineStatus(nodectx.NODE_OFFLINE)
	}
}



func (h *Handler) GetNodeInfo(c echo.Context) (err error) {
	
	// nodeopt *options.NodeOptions
	output := make(map[string]interface{})

	output[NODE_VERSION] = nodectx.GetNodeCtx().Version + " - " + h.GitCommit
	output[NODETYPE] = "peer"
	updateNodeStatus(h.Node.NetworkName)
	if nodectx.GetNodeCtx().Status == nodectx.NODE_ONLINE {
		output[NODE_STATUS] = "NODE_ONLINE"
	} else {
		output[NODE_STATUS] = "NODE_OFFLINE"
	}
	
	pubkeybytes, err := p2pcrypto.MarshalPublicKey(nodectx.GetNodeCtx().PublickKey)
	if err != nil {
		output[ERROR_INFO] = err.Error()
		return c.JSON(http.StatusBadRequest, output)
	}

	output[NODE_PUBKEY] = p2pcrypto.ConfigEncodeKey(pubkeybytes)
	output[NODE_ID] = nodectx.GetNodeCtx().PeerId.Pretty()

	peers := nodectx.GetNodeCtx().PeersProtocol()
	output[PEERS] = *peers

	return c.JSON(http.StatusOK, output)
}