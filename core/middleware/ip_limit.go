package middleware

import (
	"net"
	"strings"

	"github.com/1Panel-dev/1Panel/core/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/utils/common"
	"github.com/1Panel-dev/1Panel/core/utils/security"
	"github.com/gin-gonic/gin"
)

func WhiteAllow() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("X-Panel-Local-Token")
		clientIP := common.GetRealClientIP(c)
		if isLocalSyncRequest(c.Request.URL.Path, clientIP, tokenString) {
			c.Set("LOCAL_REQUEST", true)
			c.Next()
			return
		}
		if common.IsPrivateIP(clientIP) {
			c.Next()
			return
		}

		settingRepo := repo.NewISettingRepo()
		allowIPs, err := settingRepo.GetValueByKey("AllowIPs")
		if err != nil {
			helper.InternalServer(c, err)
			return
		}

		if len(allowIPs) == 0 {
			c.Next()
			return
		}
		for _, ip := range strings.Split(allowIPs, ",") {
			if len(ip) == 0 {
				continue
			}
			if ip == clientIP || (strings.Contains(ip, "/") && common.CheckIpInCidr(ip, clientIP)) {
				c.Next()
				return
			}
		}
		code := security.LoadErrCode()
		helper.ErrWithHtml(c, code, "err_ip_limit")
	}
}

func isLocalSyncRequest(reqPath, clientIP, token string) bool {
	ip := net.ParseIP(clientIP)
	if ip == nil || !ip.IsLoopback() {
		return false
	}

	switch reqPath {
	case "/api/v2/core/xpack/sync/ssl":
		return token != ""
	case "/api/v2/core/settings/ssl/reload":
		return true
	default:
		return false
	}
}
