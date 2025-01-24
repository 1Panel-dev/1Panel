package middleware

import (
	"github.com/1Panel-dev/1Panel/core/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/gin-gonic/gin"
)

func GlobalLoading() gin.HandlerFunc {
	return func(c *gin.Context) {
		settingRepo := repo.NewISettingRepo()
		status, err := settingRepo.Get(repo.WithByKey("SystemStatus"))
		if err != nil {
			helper.InternalServer(c, err)
			return
		}
		if status.Value != "Free" {
			helper.ErrorWithDetail(c, 407, status.Value, err)
			return
		}
		c.Next()
	}
}
