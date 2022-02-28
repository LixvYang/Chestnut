// Package api provides API for chestnut.
package api

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/lixvyang/chestnut/chain"
	chestnutpb "github.com/lixvyang/chestnut/pb"
)

type CustomValidatorPost struct {
	Validator *validator.Validate
}

type TrxResult struct {
	TrxId string `json:"trx_id"`
}

func (cv *CustomValidatorPost) Validate(i interface{}) error {
	switch i.(type) {
		case *chestnutpb.Activity:
			inputobj := i.(*chestnutpb.Activity)
			if inputobj.Type == Add {
				if inputobj.Object != nil && inputobj.Target != nil {
					if inputobj.Target.Type == Group && inputobj.Target.Id != "" {
						if inputobj.Object.Type == Note && inputobj.Object.Content != "" {
							return nil 
						}
						return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("unsupported object type: %s", inputobj.Object.Type))
					}
					return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("Target Group must not be nil"))
				}
				return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("Object and Target Object must not be nil"))
			}
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("unknown type of Actitity: %s", inputobj.Type))
		default:
			// TODO: implement
			if err := cv.Validator.Struct(i); err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
			}
	}
	return nil 
}

func (h *Handler) PostToGroup(c echo.Context) (err error) {
	output := make(map[string]string)
	paramspb := new(chestnutpb.Activity)
	if err = c.Bind(paramspb); err != nil {
		output[ERROR_INFO] = err.Error()
		return c.JSON(http.StatusBadRequest, output)
	}

	validate := &CustomValidatorPost{Validator: validator.New()}
	if err = validate.Validate(paramspb); err != nil {
		output[ERROR_INFO] = err.Error()
		return c.JSON(http.StatusBadRequest, output)
	}

	groupmgr := chain.GetGroupMgr()
	if group, ok := groupmgr.Groups[paramspb.Target.Id]; ok {
		trxId, err := group.PostToGroup(paramspb.Object)

		if err != nil {
			output[ERROR_INFO] = err.Error()
			return c.JSON(http.StatusBadRequest, output)
		}
		return c.JSON(http.StatusOK, &TrxResult{TrxId: trxId})
	} else {
		output[ERROR_INFO] = fmt.Sprintf("Group %s not exist", paramspb.Target.Id)
		return c.JSON(http.StatusBadRequest, output)
	}
}