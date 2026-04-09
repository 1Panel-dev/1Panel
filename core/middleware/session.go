package middleware

import (
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/core/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/buserr"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/utils/xpack"
	"github.com/gin-gonic/gin"
)

func SessionAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiReq := c.GetBool("API_AUTH")
		if strings.HasPrefix(c.Request.URL.Path, "/api/v2/core/auth") || apiReq {
			c.Next()
			return
		}

		psession, err := global.SESSION.Get(c)
		if err != nil {
			errItem := err.Error()
			if errItem == "ErrSessionDataFormat" || errItem == "ErrSessionDataNotFound" {
				helper.BadAuth(c, "ErrNotLogin", buserr.New(errItem))
				return
			}
			helper.BadAuth(c, "ErrNotLogin", err)
			return
		}
		if len(psession.Name) == 0 {
			helper.BadAuth(c, "ErrNotLogin", err)
			return
		}
		settingRepo := repo.NewISettingRepo()
		sessionTimeout, err := settingRepo.GetValueByKey("SessionTimeout")
		if err != nil {
			global.LOG.Errorf("create operation record failed, err: %v", err)
			return
		}
		lifeTime, _ := strconv.Atoi(sessionTimeout)
		lifeTime = xpack.LoadSessionTimeout(psession, lifeTime)
		ssl, err := settingRepo.GetValueByKey("SSL")
		if err != nil {
			global.LOG.Errorf("create operation record failed, err: %v", err)
			return
		}
		if _, err := global.SESSION.RefreshIfNeeded(c, psession, ssl == constant.StatusEnable, lifeTime); err != nil {
			errItem := err.Error()
			if errItem == "ErrSessionDataFormat" || errItem == "ErrSessionDataNotFound" {
				helper.BadAuth(c, "ErrNotLogin", buserr.New(errItem))
				return
			}
			global.LOG.Warnf("refresh session failed, path=%s, err=%v", c.Request.URL.Path, err)
			helper.BadAuth(c, "ErrNotLogin", err)
			return
		}
		c.Next()
	}
}
