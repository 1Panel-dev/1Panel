package middleware

import (
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/repo"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/gin-gonic/gin"
)

func GlobalLoading() gin.HandlerFunc {
	return func(c *gin.Context) {
		settingRepo := repo.NewISettingRepo()
		status, err := settingRepo.Get(settingRepo.WithByKey("SystemStatus"))
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
			return
		}
		if status.Value != "Free" {
			helper.ErrorWithDetail(c, constant.CodeGlobalLoading, status.Value, err)
			return
		}
		c.Next()
	}
}
