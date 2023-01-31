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
		upgradeSetting, err := settingRepo.Get(settingRepo.WithByKey("SystemVersion"))
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
			return
		}
		if upgradeSetting.Value == constant.StatusWaiting {
			helper.ErrorWithDetail(c, constant.CodeGlobalLoading, "Upgrading", err)
			return
		}
		c.Next()
	}
}
