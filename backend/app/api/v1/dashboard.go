package v1

import (
	"errors"

	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/gin-gonic/gin"
)

func (b *BaseApi) LoadDashboardBaseInfo(c *gin.Context) {
	ioOption, ok := c.Params.Get("ioOption")
	if !ok {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, errors.New("error ioOption in path"))
		return
	}
	netOption, ok := c.Params.Get("netOption")
	if !ok {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, errors.New("error netOption in path"))
		return
	}
	data, err := dashboardService.LoadBaseInfo(ioOption, netOption)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
		return
	}
	helper.SuccessWithData(c, data)
}

func (b *BaseApi) LoadDashboardCurrentInfo(c *gin.Context) {
	ioOption, ok := c.Params.Get("ioOption")
	if !ok {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, errors.New("error ioOption in path"))
		return
	}
	netOption, ok := c.Params.Get("netOption")
	if !ok {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrTypeInvalidParams, errors.New("error netOption in path"))
		return
	}
	data := dashboardService.LoadCurrentInfo(ioOption, netOption)
	helper.SuccessWithData(c, data)
}
