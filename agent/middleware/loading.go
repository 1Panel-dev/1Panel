package middleware

import (
	"net/http"

	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/gin-gonic/gin"
)

func GlobalLoading() gin.HandlerFunc {
	return func(c *gin.Context) {
		settingRepo := repo.NewISettingRepo()
		status, err := settingRepo.Get(settingRepo.WithByKey("SystemStatus"))
		if err != nil {
			helper.InternalServer(c, err)
			return
		}
		if status.Value != "Free" {
			helper.ErrorWithDetail(c, http.StatusProxyAuthRequired, status.Value, err)
			return
		}
		c.Next()
	}
}
