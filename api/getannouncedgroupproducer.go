// Package api provides API for chestnut.
package api

import (
	"net/http"
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/lixvyang/chestnut/chain"
)

type AnnouncedProducerListItem struct {
	AnnouncedPubkey string
	AnnouncerSign   string
	Result          string
	TimeStamp       int64
}

func (h *Handler) GetAnnouncedGroupProducer(c echo.Context) (err error) {
	output := make(map[string]string)
	groupid := c.Param("group_id")
	if groupid == "" {
		output[ERROR_INFO] = "group_id can't be nil."
		return c.JSON(http.StatusBadRequest, output)
	}
	groupmgr := chain.GetGroupMgr()
	if group, ok := groupmgr.Groups[groupid]; ok {
		prdList, err := group.GetAnnouncedProducer()
		if err != nil {
			output[ERROR_INFO] = err.Error()
			return c.JSON(http.StatusBadRequest, output)
		}

		prdResultList := []*AnnouncedProducerListItem{}
		for _, prd := range prdList {
			var item *AnnouncedProducerListItem
			item = &AnnouncedProducerListItem{}
			item.AnnouncedPubkey = prd.SignPubkey
			item.AnnouncerSign = prd.AnnouncerSignature
			item.Result = prd.Result.String()
			item.TimeStamp = prd.TimeStamp
			prdResultList = append(prdResultList, item)
		}

		return c.JSON(http.StatusOK, prdResultList)
	} else {
		output[ERROR_INFO] = fmt.Sprintf("Group %s not exist", groupid)
		return c.JSON(http.StatusBadRequest, output)
	}


}