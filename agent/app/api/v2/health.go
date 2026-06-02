package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/gin-gonic/gin"
)

// @Tags Health
// @Summary Check health
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /health/check [get]
func (b *BaseApi) CheckHealth(c *gin.Context) {
	helper.Success(c)
}
