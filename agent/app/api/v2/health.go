package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/gin-gonic/gin"
)

func (b *BaseApi) CheckHealth(c *gin.Context) {
	helper.Success(c)
}
