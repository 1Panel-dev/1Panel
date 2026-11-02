package middleware

import (
	"errors"

	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/global"
	contract "github.com/1Panel-dev/1Panel/core/utils/mcp"
	"github.com/1Panel-dev/1Panel/core/utils/mcpapi"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func MCPSourcePolicy() SourcePolicy {
	return SourcePolicy{
		Bypass: func(c *gin.Context) bool {
			return mcpapi.IsRequest(c.Request) && c.GetBool("MCP_API_AUTH")
		},
		Resolve: resolveMCPSource,
	}
}

func resolveMCPSource(c *gin.Context, trustedProxies string) (string, bool) {
	clientIP, err := resolveMCPWebsiteClientIP(c, trustedProxies)
	if err != nil {
		rejectMCPWebsiteSource(c, err)
		return "", false
	}
	if c.Request.URL.Path != "/mcp" {
		return clientIP, false
	}
	c.Set("MCP_CLIENT_IP", clientIP)
	var policy model.MCPSetting
	err = global.DB.Select("access_policy", "allowed_ips").First(&policy, 1).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		c.AbortWithStatusJSON(503, gin.H{"code": "MCP_POLICY_UNAVAILABLE", "message": "MCP source policy is unavailable"})
		return "", false
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		policy.AccessPolicy = "inherit"
	}
	switch policy.AccessPolicy {
	case "inherit":
		return clientIP, false
	case "all":
		return clientIP, true
	case "restricted":
		entries, err := contract.DecodeAllowedIPs(policy.AllowedIPs)
		if err == nil && len(entries) > 0 && contract.IPAllowed(entries, clientIP) {
			return clientIP, true
		}
		c.AbortWithStatusJSON(403, gin.H{"code": "MCP_SOURCE_DENIED", "message": "Source IP is not allowed by MCP policy"})
	default:
		c.AbortWithStatusJSON(403, gin.H{"code": "MCP_SOURCE_DENIED", "message": "MCP source policy is invalid"})
	}
	return "", false
}
