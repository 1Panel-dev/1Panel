package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/gin-gonic/gin"
)

func (b *BaseApi) TerminalCapabilities(c *gin.Context) {
	helper.SuccessWithData(c, gin.H{"apiKeyLeaseVersion": 1})
}
