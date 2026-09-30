package middleware

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func WebSocketOriginGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !websocket.IsWebSocketUpgrade(c.Request) {
			c.Next()
			return
		}

		if site := c.GetHeader("Sec-Fetch-Site"); site != "" && site != "same-origin" {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		origins := c.Request.Header.Values("Origin")
		if len(origins) == 0 {
			c.Next()
			return
		}
		if len(origins) != 1 {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		origin, err := url.Parse(origins[0])
		if err != nil || (origin.Scheme != "http" && origin.Scheme != "https") ||
			origin.Host == "" || !strings.EqualFold(origin.Host, c.Request.Host) ||
			origin.User != nil || origin.Path != "" || origin.RawQuery != "" || origin.ForceQuery || origin.Fragment != "" {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.Next()
	}
}
