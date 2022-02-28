// Package api provides API for chestnut.
package api

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/lixvyang/chestnut/chain"
)


type ProducerListItem struct {
	ProducerPubkey string
	OwnerPubkey    string
	OwnerSign      string
	TimeStamp      int64
	BlockProduced  int64
}


func (h *Handler) GetGroupProducers(c echo.Context) (err error) {
	output := make(map[string]string)
	groupid := c.Param("group_id")

	if groupid == "" {
		output[ERROR_INFO] = "group_id can't be nil."
		return c.JSON(http.StatusBadRequest, output)
	}

	groupmgr := chain.GetGroupMgr()
	if group, ok := groupmgr.Groups[groupid]; ok {
		prdList, err := group.GetProducers()
		if err != nil {
			output[ERROR_INFO] = err.Error()
			return c.JSON(http.StatusBadRequest, output)
		}

		var prdResultList []*ProducerListItem
		for _, prd := range prdList {
			item := &ProducerListItem{}
			item.ProducerPubkey = prd.ProducerPubkey
			item.OwnerPubkey = prd.GroupOwnerPubkey
			item.OwnerSign = prd.GroupOwnerSign
			item.TimeStamp = prd.TimeStamp
			item.BlockProduced = prd.BlockProduced
			prdResultList = append(prdResultList, item)
		}
		return c.JSON(http.StatusOK, prdResultList)
	} else {
		output[ERROR_INFO] = fmt.Sprintf("Group %s not exist", groupid)
		return c.JSON(http.StatusBadRequest, output)
	}
}