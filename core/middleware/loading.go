package middleware

import (
	"github.com/1Panel-dev/1Panel/core/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/gin-gonic/gin"
)

func GlobalLoading() gin.HandlerFunc {
	return func(c *gin.Context) {
		settingRepo := repo.NewISettingRepo()
		commonRepo := repo.NewICommonRepo()
		status, err := settingRepo.Get(commonRepo.WithByKey("SystemStatus"))
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
