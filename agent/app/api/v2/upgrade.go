package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/gin-gonic/gin"
)

func (b *BaseApi) Upgrade(c *gin.Context) {
	var req dto.Upgrade
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := upgradeService.Upgrade(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, nil)
}

func (b *BaseApi) Rollback(c *gin.Context) {
	var req dto.Rollback
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := upgradeService.Rollback(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, nil)
}
