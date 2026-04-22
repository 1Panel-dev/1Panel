package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/core/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/utils/common"
	"github.com/gin-gonic/gin"
)

var (
	expiredLoc     *time.Location
	expiredLocOnce sync.Once
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
		expirationDays, err := settingRepo.GetValueByKey("ExpirationDays")
		if err != nil {
			helper.ErrorWithDetail(c, http.StatusInternalServerError, "ErrPasswordExpired", err)
			return
		}
		expiredDays, _ := strconv.Atoi(expirationDays)
		if expiredDays == 0 {
			c.Next()
			return
		}

		expirationTime, err := settingRepo.GetValueByKey("ExpirationTime")
		if err != nil {
			helper.ErrorWithDetail(c, http.StatusInternalServerError, "ErrPasswordExpired", err)
			return
		}
		expiredTime, err := time.ParseInLocation(constant.DateTimeLayout, expirationTime, loadExpiredLocation())
		if err != nil {
			helper.ErrorWithDetail(c, 313, "ErrPasswordExpired", err)
			return
		}
		if time.Now().After(expiredTime) {
			helper.ErrorWithDetail(c, 313, "ErrPasswordExpired", err)
			return
		}
		c.Next()
	}
}

func loadExpiredLocation() *time.Location {
	expiredLocOnce.Do(func() {
		loc, err := time.LoadLocation(common.LoadTimeZoneByCmd())
		if err != nil {
			expiredLoc = time.Local
			return
		}
		expiredLoc = loc
	})
	if expiredLoc == nil {
		return time.Local
	}
	return expiredLoc
}
