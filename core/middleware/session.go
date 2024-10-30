package middleware

import (
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/core/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/gin-gonic/gin"
)

func SessionAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/v2/core/auth") {
			c.Next()
			return
		}
		if method, exist := c.Get("authMethod"); exist && method == constant.AuthMethodJWT {
			c.Next()
			return
		}
		sId, err := c.Cookie(constant.SessionName)
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrUnauthorized, constant.ErrTypeNotLogin, nil)
			return
		}
		psession, err := global.SESSION.Get(sId)
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrUnauthorized, constant.ErrTypeNotLogin, nil)
			return
		}
		settingRepo := repo.NewISettingRepo()
		commonRepo := repo.NewICommonRepo()
		setting, err := settingRepo.Get(commonRepo.WithByKey("SessionTimeout"))
		if err != nil {
			global.LOG.Errorf("create operation record failed, err: %v", err)
		}
		lifeTime, _ := strconv.Atoi(setting.Value)
		_ = global.SESSION.Set(sId, psession, lifeTime)
		c.Next()
	}
}
