package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/core/app/api/v2/helper"
	baseRepo "github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/buserr"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	psessionUtils "github.com/1Panel-dev/1Panel/core/init/session/psession"
	"github.com/1Panel-dev/1Panel/core/utils/common"
	"github.com/1Panel-dev/1Panel/core/utils/xpack"
	"github.com/gin-gonic/gin"
)

func PasswordExpired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, "/api/v2/") {
			c.Next()
			return
		}
		if strings.HasPrefix(c.Request.URL.Path, "/api/v2/core/auth") ||
			c.Request.URL.Path == "/api/v2/core/settings/search" ||
			c.Request.URL.Path == "/api/v2/core/settings/search/base" ||
			c.Request.URL.Path == "/api/v2/core/xpack/settings/search" ||
			c.Request.URL.Path == "/api/v2/core/xapp/verifyQRCode" ||
			c.Request.URL.Path == "/api/v2/core/enterprise/users/info" ||
			c.Request.URL.Path == "/api/v2/core/enterprise/licenses/info" ||
			c.Request.URL.Path == "/api/v2/core/enterprise/licenses/status" ||
			c.Request.URL.Path == "/api/v2/core/enterprise/licenses/upload" {
			c.Next()
			return
		}
		if c.GetBool("API_AUTH") {
			c.Next()
			return
		}
		var err error
		sessionUser, ok := c.Get(psessionUtils.GinContextSessionUserKey)
		psession, typeOK := sessionUser.(psessionUtils.SessionUser)
		if !ok || !typeOK {
			psession, err = global.SESSION.Get(c)
		}
		if err != nil {
			errItem := err.Error()
			if errItem == "ErrSessionDataFormat" || errItem == "ErrSessionDataNotFound" {
				helper.BadAuth(c, "ErrNotLogin", buserr.New(errItem))
				return
			}
			helper.BadAuth(c, "ErrNotLogin", err)
			return
		}
		c.Set(psessionUtils.GinContextSessionUserKey, psession)
		if len(psession.Name) == 0 {
			helper.BadAuth(c, "ErrNotLogin", err)
			return
		}
		settingRepo := baseRepo.NewISettingRepo()
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
		expirationTime, err := xpack.AuthProvider.LoadPasswordExpirationTime(c)
		if err != nil {
			helper.ErrorWithDetail(c, http.StatusInternalServerError, "ErrPasswordExpired", err)
			return
		}
		expiredTime, err := time.ParseInLocation(constant.DateTimeLayout, expirationTime, common.LoadExpiredLocation())
		if err != nil {
			helper.ErrorWithDetail(c, http.StatusInternalServerError, "ErrPasswordExpired", err)
			return
		}

		if time.Now().After(expiredTime) {
			helper.ErrorWithDetail(c, 313, "ErrPasswordExpired", err)
			return
		}
		c.Next()
	}
}
