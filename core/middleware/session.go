package middleware

import (
	"github.com/1Panel-dev/1Panel/core/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
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
			helper.BadAuth(c, "ErrNotLogin", err)
			return
		}
		settingRepo := repo.NewISettingRepo()
		setting, err := settingRepo.Get(repo.WithByKey("SessionTimeout"))
		if err != nil {
			global.LOG.Errorf("create operation record failed, err: %v", err)
			return
		}
		lifeTime, _ := strconv.Atoi(setting.Value)
		httpsSetting, err := settingRepo.Get(repo.WithByKey("SSL"))
		if err != nil {
			global.LOG.Errorf("create operation record failed, err: %v", err)
			return
		}
		_ = global.SESSION.Set(c, psession, httpsSetting.Value == constant.StatusEnable, lifeTime)
		c.Next()
	}
}
