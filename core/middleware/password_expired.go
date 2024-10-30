package middleware

import (
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/core/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/utils/common"
	"github.com/gin-gonic/gin"
)

func PasswordExpired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/v2/core/auth") ||
			c.Request.URL.Path == "/api/v2/core/settings/expired/handle" ||
			c.Request.URL.Path == "/api/v2/core/settings/search" {
			c.Next()
			return
		}
		settingRepo := repo.NewISettingRepo()
		commonRepo := repo.NewICommonRepo()
		setting, err := settingRepo.Get(commonRepo.WithByKey("ExpirationDays"))
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypePasswordExpired, err)
			return
		}
		expiredDays, _ := strconv.Atoi(setting.Value)
		if expiredDays == 0 {
			c.Next()
			return
		}

		extime, err := settingRepo.Get(commonRepo.WithByKey("ExpirationTime"))
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypePasswordExpired, err)
			return
		}
		loc, _ := time.LoadLocation(common.LoadTimeZone())
		expiredTime, err := time.ParseInLocation(constant.DateTimeLayout, extime.Value, loc)
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodePasswordExpired, constant.ErrTypePasswordExpired, err)
			return
		}
		if time.Now().After(expiredTime) {
			helper.ErrorWithDetail(c, constant.CodePasswordExpired, constant.ErrTypePasswordExpired, err)
			return
		}
		c.Next()
	}
}
