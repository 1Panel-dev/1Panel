package middleware

import (
	"crypto/hmac"
	"errors"
	"net"
	"strings"

	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/utils/common"
	"github.com/1Panel-dev/1Panel/core/utils/encrypt"
	contract "github.com/1Panel-dev/1Panel/core/utils/mcp"
	"github.com/gin-gonic/gin"
)

func resolveMCPWebsiteClientIP(c *gin.Context, trustedProxies string) (string, error) {
	keys := c.Request.Header.Values(contract.WebsiteProxyKeyHeader)
	peers := c.Request.Header.Values(contract.WebsiteProxyPeerHeader)
	c.Request.Header.Del(contract.WebsiteProxyKeyHeader)
	c.Request.Header.Del(contract.WebsiteProxyPeerHeader)
	if c.Request.URL.Path != "/mcp" || len(keys) == 0 && len(peers) == 0 {
		return common.ResolveClientIP(c, trustedProxies), nil
	}
	if len(keys) != 1 || len(peers) != 1 || !isLocalMCPWebsitePeer(common.GetRealClientIP(c)) {
		return "", contract.Forbidden()
	}
	peer := net.ParseIP(strings.TrimSpace(peers[0]))
	if peer == nil {
		return "", contract.Forbidden()
	}
	if global.DB == nil {
		return "", contract.Fail("MCP_POLICY_UNAVAILABLE", 503, "MCP source policy is unavailable")
	}
	var setting model.MCPSetting
	if err := global.DB.Select("website_id", "website_binding_id").First(&setting, 1).Error; err != nil {
		return "", contract.Fail("MCP_POLICY_UNAVAILABLE", 503, "MCP source policy is unavailable")
	}
	if setting.WebsiteID == 0 {
		return "", contract.Forbidden()
	}
	masterKey, err := encrypt.APIKeyEncryptionKey()
	if err != nil {
		return "", contract.Fail("MCP_POLICY_UNAVAILABLE", 503, "MCP source policy is unavailable")
	}
	expected, err := contract.WebsiteProxyKey(masterKey, setting.WebsiteBindingID)
	if err != nil || !hmac.Equal([]byte(expected), []byte(keys[0])) {
		return "", contract.Forbidden()
	}
	request := c.Request.Clone(c.Request.Context())
	request.RemoteAddr = net.JoinHostPort(peer.String(), "0")
	proxyContext := &gin.Context{Request: request}
	c.Set("MCP_WEBSITE_PROXY", true)
	return common.ResolveClientIP(proxyContext, trustedProxies), nil
}

func isLocalMCPWebsitePeer(value string) bool {
	peer := net.ParseIP(value)
	if peer == nil {
		return false
	}
	if peer.IsLoopback() {
		return true
	}
	bound := net.ParseIP(strings.Trim(global.CONF.Conn.BindAddress, "[]"))
	return bound != nil && !bound.IsUnspecified() && bound.Equal(peer)
}

func rejectMCPWebsiteSource(c *gin.Context, err error) {
	var known *contract.Error
	if errors.As(err, &known) && known.Status == 503 {
		c.AbortWithStatusJSON(503, gin.H{"code": "MCP_POLICY_UNAVAILABLE", "message": "MCP source policy is unavailable"})
		return
	}
	c.AbortWithStatusJSON(403, gin.H{"code": "MCP_SOURCE_DENIED", "message": "MCP website proxy identity is invalid"})
}
