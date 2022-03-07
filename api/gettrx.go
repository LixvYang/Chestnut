// Package api provides API for chestnut.
package api

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/lixvyang/chestnut/chain"
	_ "github.com/lixvyang/chestnut/pb"
)

type GetTrxParam struct {
	TrxId string `from:"trx_id" json:"trx_id" validate:"required"`
}

func (h *Handler) GetTrx(c echo.Context) (err error) {
	output := make(map[string]string)
	groupid := c.Param("group_id")
	if groupid == "" {
		output[ERROR_INFO] = "group_id can't be nil."
		return c.JSON(http.StatusBadRequest, output)
	}

	trxid := c.Param("trx_id")
	if trxid == "" {
		output[ERROR_INFO] = "trx_id can't be nil."
		return c.JSON(http.StatusBadRequest, output)
	}

	groupmgr := chain.GetGroupMgr()
	if group, ok := groupmgr.Groups[groupid]; ok {
		trx, err := group.GetTrx(trxid)
		if err != nil {
			output[ERROR_INFO] = err.Error()
			return c.JSON(http.StatusBadRequest, output)
		}
		return c.JSON(http.StatusOK, trx)
	} else {
		output[ERROR_INFO] = fmt.Sprintf("Group %s not exist", groupid)
		return c.JSON(http.StatusBadRequest, output)
	}
}