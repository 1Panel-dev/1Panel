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

type SourcePolicy struct {
	Bypass  func(*gin.Context) bool
	Resolve func(*gin.Context, string) (clientIP string, allowed bool)
}

func WhiteAllow(policies ...SourcePolicy) gin.HandlerFunc {
	policies = append([]SourcePolicy(nil), policies...)
	return func(c *gin.Context) {
		for _, policy := range policies {
			if policy.Bypass == nil {
				continue
			}
			allowed := policy.Bypass(c)
			if c.IsAborted() {
				return
			}
			if allowed {
				c.Next()
				return
			}
		}
		tokenString := c.GetHeader("X-Panel-Local-Token")
		remoteIP := common.GetRealClientIP(c)
		if isLocalSyncRequest(c.Request.URL.Path, remoteIP, tokenString) {
			c.Set("LOCAL_REQUEST", true)
			c.Next()
			return
		}

		settingRepo := repo.NewISettingRepo()
		trustedProxies, err := settingRepo.GetValueByKey("AllowIPTrustedProxies")
		if err != nil {
			helper.InternalServer(c, err)
			return
		}
		clientIP := ""
		for _, policy := range policies {
			if policy.Resolve == nil {
				continue
			}
			resolved, allowed := policy.Resolve(c, trustedProxies)
			if c.IsAborted() {
				return
			}
			if resolved != "" {
				clientIP = resolved
			}
			if allowed {
				c.Next()
				return
			}
		}
		if clientIP == "" {
			clientIP = common.ResolveClientIP(c, trustedProxies)
		}

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
	case "/api/v2/core/xpack/sync/ssl", "/api/v2/core/xpack/sync/php/extensions":
		return token != ""
	case "/api/v2/core/settings/ssl/reload":
		return token != ""
	default:
		return false
	}
}
