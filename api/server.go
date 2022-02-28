// Package api provides API for chestnut
package api

import (
	"fmt"
	"io/ioutil"
	"os"
	"syscall"

	"github.com/labstack/echo/v4"
	localcrypto "github.com/lixvyang/chestnut/crypto"
	"github.com/lixvyang/chestnut/p2p"
	chestnutpb "github.com/lixvyang/chestnut/pb"
	appapi "github.com/lixvyang/chestnut/pkg/app/api"
	"github.com/lixvyang/chestnut/utils/cli"
	"github.com/lixvyang/chestnut/utils/options"
	"google.golang.org/protobuf/encoding/protojson"
)

var quitch chan os.Signal

func StartAPIServer(config cli.Config, signalch chan os.Signal,h *Handler, apph *appapi.Handler, node *p2p.Node, nodeopt *options.NodeOptions, ks localcrypto.Keystore, ethaddr string, isbootstrapnode bool)  {
	quitch = signalch
	e := echo.New()
	e.Binder = new(CustomBinder)
	r := e.Group("api")
	a := e.Group("app/api")
	r.GET("/quit", quitapp)
	if !isbootstrapnode {
		r.GET("v1/node", h.GetNodeInfo)
		r.POST("v1/group", h.CreateGroup)
		r.DELETE("v1/group", h.RmGroup)
		r.POST("v1/group/join", h.JoinGroup)
		r.POST("v1/group/leave", h.LeaveGroup)
		r.POST("v1/group/content", h.PostToGroup)
		r.POST("v1/group/profile", h.UpdateProfile)
		r.POST("v1/network/peers", h.AddPeers)	
		r.POST("/v1/group/deniedlist", h.MgrGrpBlkList)
		r.POST("v1/group/producer", h.GroupProducer)
		r.POST("v1/group/announce", h.Announce)
		r.POST("/v1/group/schema", h.Schema)
		r.POST("/v1/group/:group_id/startsync", h.StartSync)
		r.GET("v1/network", h.GetNetwork(&node.Host, node.Info, nodeopt, ethaddr))
		r.POST("/v1/psping", h.PSPingPeer(node))
		r.GET("/v1/block/:group_id/:block_id", h.GetBlockById)
		r.GET("/v1/trx/:group_id/:trx_id", h.GetTrx)
		r.GET("/v1/groups", h.GetGroups)
		r.GET("/v1/group/:group_id/content", h.GetGroupCtn)
		r.GET("/v1/group/:group_id/deniedlist", h.GetDeniedUserList)
		r.GET("/v1/group/:group_id/producers", h.GetGroupProducers)
		r.GET("/v1/group/:group_id/announced/users", h.GetAnnouncedGroupUsers)
		r.GET("/v1/group/:group_id/announced/producers", h.GetAnnouncedGroupProducer)
		r.GET("/v1/group/:group_id/app/schema", h.GetGroupAppSchema)

		a.POST("/v1/group/:group_id/content", apph.ContentByPeers)
		// a.POST("/v1/token/apply", apph.ApplyToken)
		// a.POST("/v1/token/refresh", apph.RefreshToken)

	}

	e.Logger.Fatal(e.Start(config.APIListenAddresses))
}



type CustomBinder struct{}

func (cb *CustomBinder) Bind(i interface{}, c echo.Context) (err error) {
	db := new(echo.DefaultBinder)
	switch i.(type) {
	case *chestnutpb.Activity:
		bodyBytes, err := ioutil.ReadAll(c.Request().Body)
		err = protojson.Unmarshal(bodyBytes, i.(*chestnutpb.Activity))
		return err
	default:
		if err = db.Bind(i, c); err != echo.ErrUnsupportedMediaType {
			return 
		}
		return err
	}
}

func quitapp(c echo.Context) (err error) {
	fmt.Println("/api/quit has been called, send Signal SIGNERM...")
	quitch <- syscall.SIGTERM
	return nil
}