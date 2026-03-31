package middleware

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/gin-gonic/gin"
)

func CSRFSameOrigin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !requiresCSRFSameOriginCheck(c) {
			c.Next()
			return
		}

		if !isTrustedRequestSource(c) {
			c.AbortWithStatusJSON(http.StatusForbidden, dto.Response{
				Code:    http.StatusForbidden,
				Message: "invalid request origin",
				Data:    nil,
			})
			return
		}

		c.Next()
	}
}

func requiresCSRFSameOriginCheck(c *gin.Context) bool {
	unsafeMethod := c.Request.Method != http.MethodGet &&
		c.Request.Method != http.MethodHead &&
		c.Request.Method != http.MethodOptions &&
		c.Request.Method != http.MethodTrace
	if !unsafeMethod {
		return false
	}
	if !strings.HasPrefix(c.Request.URL.Path, "/api/v2/") {
		return false
	}
	if strings.HasPrefix(c.Request.URL.Path, "/api/v2/core/auth") {
		return false
	}
	if c.GetBool("API_AUTH") {
		return false
	}
	sessionID, err := c.Cookie(constant.SessionName)
	return err == nil && sessionID != ""
}

func isTrustedRequestSource(c *gin.Context) bool {
	origin := strings.TrimSpace(c.GetHeader("Origin"))
	if origin != "" {
		return sameRequestHost(c, origin)
	}

	referer := strings.TrimSpace(c.GetHeader("Referer"))
	if referer != "" {
		return sameRequestHost(c, referer)
	}

	return false
}

func sameRequestHost(c *gin.Context, rawURL string) bool {
	requestScheme := requestScheme(c)
	requestHost := strings.TrimSpace(c.Request.Host)
	if requestHost != "" && sameOrigin(rawURL, requestScheme, requestHost) {
		return true
	}

	for _, header := range []string{"X-Forwarded-Host", "X-Original-Host"} {
		value := strings.TrimSpace(c.GetHeader(header))
		if value == "" {
			continue
		}
		parts := strings.Split(value, ",")
		if len(parts) == 0 {
			continue
		}
		host := strings.TrimSpace(parts[0])
		if host != "" && sameOrigin(rawURL, requestScheme, host) {
			return true
		}
	}

	return false
}

func requestScheme(c *gin.Context) string {
	if c.Request.TLS != nil {
		return "https"
	}
	if proto := forwardedProto(c.GetHeader("X-Forwarded-Proto")); proto != "" {
		return proto
	}
	return "http"
}

func forwardedProto(value string) string {
	if value == "" {
		return ""
	}
	parts := strings.Split(value, ",")
	if len(parts) == 0 {
		return ""
	}
	proto := strings.ToLower(strings.TrimSpace(parts[0]))
	switch proto {
	case "http", "https":
		return proto
	default:
		return ""
	}
}

func sameOrigin(rawURL, requestScheme, requestHost string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return false
	}
	return strings.EqualFold(parsed.Scheme, requestScheme) && strings.EqualFold(parsed.Host, requestHost)
}
