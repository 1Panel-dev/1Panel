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
		tokenString := c.GetHeader("X-Panel-Local-Token")
		clientIP := common.GetRealClientIP(c)
		if clientIP == "127.0.0.1" && tokenString != "" && c.Request.URL.Path == "/api/v2/core/xpack/sync/ssl" {
			c.Set("LOCAL_REQUEST", true)
			c.Next()
			return
		}
		if common.IsPrivateIP(clientIP) {
			c.Next()
			return
		}

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
