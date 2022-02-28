// Package api provides API for chestnut.
package api

import (
	"net/http"
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/lixvyang/chestnut/chain"
)

type DeniedUserListItem struct {
	GroupId          string
	PeerId           string
	GroupOwnerPubkey string
	GroupOwnerSign   string
	TimeStamp        int64
	Action           string
	Memo             string
}

func (h *Handler) GetDeniedUserList(c echo.Context) (err error) {
	output := make(map[string]string)
	var result []*DeniedUserListItem

	groupid := c.Param("group_id")
	if groupid == "" {
		output[ERROR_INFO] = "group_id can't be nil."
		return c.JSON(http.StatusBadRequest, output)
	}

	groupmgr := chain.GetGroupMgr()
	if group, ok := groupmgr.Groups[groupid]; ok {
		blkList, err := group.GetBlockedUser()
		if err != nil {
			output[ERROR_INFO] = err.Error()
			return c.JSON(http.StatusBadRequest, output)
		}
		for _, blkItem := range blkList {
			var item *DeniedUserListItem
			item = &DeniedUserListItem{}

			item.GroupId = blkItem.GroupId
			item.PeerId = blkItem.PeerId
			item.GroupOwnerPubkey = blkItem.GroupOwnerPubkey
			item.GroupOwnerSign = blkItem.GroupOwnerSign
			item.Action = blkItem.Action
			item.Memo = blkItem.Memo
			item.TimeStamp = blkItem.TimeStamp
			result = append(result, item)
		}
		return c.JSON(http.StatusOK, result)
	} else {
		output[ERROR_INFO] = fmt.Sprintf("Group %s not exist", groupid)
		return c.JSON(http.StatusBadRequest, output)
	}

}
