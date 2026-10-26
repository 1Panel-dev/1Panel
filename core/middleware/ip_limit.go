package middleware

import (
	"errors"
	"net"
	"strings"

	"github.com/1Panel-dev/1Panel/core/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/utils/common"
	contract "github.com/1Panel-dev/1Panel/core/utils/mcp"
	"github.com/1Panel-dev/1Panel/core/utils/mcpapi"
	"github.com/1Panel-dev/1Panel/core/utils/security"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func WhiteAllow() gin.HandlerFunc {
	return func(c *gin.Context) {
		if mcpapi.IsRequest(c.Request) && c.GetBool("MCP_API_AUTH") {
			c.Next()
			return
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
		clientIP := common.ResolveClientIP(c, trustedProxies)
		if c.Request.URL.Path == "/mcp" {
			c.Set("MCP_CLIENT_IP", clientIP)
			var policy model.MCPSetting
			err := global.DB.Select("access_policy", "allowed_ips").First(&policy, 1).Error
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				c.AbortWithStatusJSON(503, gin.H{"code": "MCP_POLICY_UNAVAILABLE", "message": "MCP source policy is unavailable"})
				return
			}
			if errors.Is(err, gorm.ErrRecordNotFound) {
				policy.AccessPolicy = "inherit"
			}
			switch policy.AccessPolicy {
			case "inherit":
			case "all":
				c.Next()
				return
			case "restricted":
				entries, err := contract.DecodeAllowedIPs(policy.AllowedIPs)
				if err == nil && len(entries) > 0 && contract.IPAllowed(entries, clientIP) {
					c.Next()
					return
				}
				c.AbortWithStatusJSON(403, gin.H{"code": "MCP_SOURCE_DENIED", "message": "Source IP is not allowed by MCP policy"})
				return
			default:
				c.AbortWithStatusJSON(403, gin.H{"code": "MCP_SOURCE_DENIED", "message": "MCP source policy is invalid"})
				return
			}
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
	case "/api/v2/core/xpack/sync/ssl":
		return token != ""
	case "/api/v2/core/settings/ssl/reload":
		return token != ""
	default:
		return false
	}
}
