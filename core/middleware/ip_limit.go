package middleware

import (
	"github.com/1Panel-dev/1Panel/core/utils/common"
	"strings"

	"github.com/1Panel-dev/1Panel/core/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/gin-gonic/gin"
)

func WhiteAllow() gin.HandlerFunc {
	return func(c *gin.Context) {
		settingRepo := repo.NewISettingRepo()
		status, err := settingRepo.Get(repo.WithByKey("AllowIPs"))
		if err != nil {
			helper.InternalServer(c, err)
			return
		}

		if len(status.Value) == 0 {
			c.Next()
			return
		}
		clientIP := c.ClientIP()
		for _, ip := range strings.Split(status.Value, ",") {
			if len(ip) == 0 {
				continue
			}
			if ip == clientIP || (strings.Contains(ip, "/") && common.CheckIpInCidr(ip, clientIP)) {
				c.Next()
				return
			}
		}
		code := LoadErrCode()
		helper.ErrWithHtml(c, code, "err_ip_limit")
	}
}
