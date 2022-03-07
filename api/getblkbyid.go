// Package api provides API for chestnut.
package api

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/lixvyang/chestnut/chain"
)


func (h *Handler) GetBlockById(c echo.Context) (err error) {
	output := make(map[string]string)
	groupid := c.Param("group_id")
	if groupid == "" {
		output[ERROR_INFO] = "group_id can't be nil."
		return c.JSON(http.StatusBadRequest, output)
	}

	blockid := c.Param("block_id")
	if blockid == "" {
		output[ERROR_INFO] = "block_id can't be nil."
		return c.JSON(http.StatusBadRequest, output)
	}

	groupmgr := chain.GetGroupMgr()
	if group, ok := groupmgr.Groups[groupid]; ok {
		block, err := group.GetBlock(blockid)
		if err != nil {
			output[ERROR_INFO] = err.Error()
			return c.JSON(http.StatusBadRequest, output)
		}
		return c.JSON(http.StatusBadRequest, block)
	} else {
		output[ERROR_INFO] = fmt.Sprintf("Group %s not exist", groupid)
		return c.JSON(http.StatusBadRequest, output)
	}
} 