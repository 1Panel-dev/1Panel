package middleware

import (
	"fmt"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

func Proxy() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/v1/auth") ||
			strings.HasPrefix(c.Request.URL.Path, "/api/v1/setting") ||
			strings.HasPrefix(c.Request.URL.Path, "/api/v1/log") {
			c.Next()
			return
		}
		target, err := url.Parse("http://127.0.0.1:9998")
		if err != nil {
			fmt.Printf("Failed to parse target URL: %v", err)
		}
		proxy := httputil.NewSingleHostReverseProxy(target)
		c.Request.Host = target.Host
		c.Request.URL.Scheme = target.Scheme
		c.Request.URL.Host = target.Host
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
