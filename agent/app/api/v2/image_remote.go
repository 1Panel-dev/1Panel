package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/service"
	"github.com/gin-gonic/gin"
)

func (b *BaseApi) GetImageRemoteCheckStatus(c *gin.Context) {
	status, err := service.GetImageRemoteCheckStatus()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, status)
}
