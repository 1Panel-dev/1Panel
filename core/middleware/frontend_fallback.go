package middleware

import (
	"net/http"
	"strings"

	"github.com/1Panel-dev/1Panel/core/utils/security"
	"github.com/gin-gonic/gin"
)

func FrontendFallback() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isFrontendFallbackRequest(c) {
			c.Next()
			return
		}
		if security.CanServeFrontendPath(c) {
			security.ToIndexHtml(c)
			c.Abort()
			return
		}
		security.HandleNotSecurity(c, "")
		c.Abort()
	}
}

func isFrontendFallbackRequest(c *gin.Context) bool {
	if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
		return false
	}
	path := c.Request.URL.Path
	if strings.HasPrefix(path, "/api/") ||
		strings.HasPrefix(path, "/assets/") ||
		strings.HasPrefix(path, "/public") ||
		strings.HasPrefix(path, "/1panel/swagger/") ||
		path == "/favicon.ico" {
		return false
	}
	if !security.IsFrontendPath(path) {
		return false
	}
	accept := c.GetHeader("Accept")
	return accept == "" || strings.Contains(accept, "text/html")
}
