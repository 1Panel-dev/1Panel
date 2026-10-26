package middleware

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/1Panel-dev/1Panel/core/utils/mcpapi"
	"github.com/1Panel-dev/1Panel/core/utils/xpack"
	"github.com/1Panel-dev/1Panel/core/utils/xpack/providers"
	"github.com/gin-gonic/gin"
)

func MCPAPIIdentity() gin.HandlerFunc {
	return func(c *gin.Context) {
		mcpapi.ClearForwardedHeaders(c.Request.Header)
		_, owner, ok := mcpapi.Identity(c.Request)
		if !ok {
			if mcpapi.HasIdentity(c.Request) {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			c.Next()
			return
		}
		provider, supported := xpack.AuthProvider.(providers.MCPAPIAuthorizer)
		nodeID, _ := mcpapi.NodeID(c.Request)
		if !supported || provider.AuthorizeMCPAPI(owner, nodeID, c.Request.Method, c.Request.URL.Path) != nil || provider.BindMCPAPIIdentity(c, owner) != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": http.StatusForbidden, "message": "MCP API permission denied"})
			return
		}
		nodeProvider, supported := xpack.AuthProvider.(providers.MCPNodeProvider)
		if !supported {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		node, err := nodeProvider.ResolveMCPNode(owner, nodeID)
		if err != nil || node.ID != nodeID || node.IsLocal != (nodeID == "local") || node.IsLocal != (node.Name == "local") || node.Name == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": http.StatusForbidden, "message": "MCP node permission denied"})
			return
		}
		c.Request.Header.Set("CurrentNode", url.QueryEscape(node.Name))
		c.Next()
	}
}

func MCPTokenBoundary() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if path == "/mcp" {
			c.Next()
			return
		}
		for _, value := range c.Request.Header.Values("Authorization") {
			for _, part := range strings.Fields(value) {
				if strings.HasPrefix(part, "mcp_") {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "MCP_UNAUTHENTICATED", "message": "MCP credentials cannot access this endpoint"})
					return
				}
			}
		}
		c.Next()
	}
}
