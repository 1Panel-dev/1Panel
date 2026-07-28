package middleware

import (
	"net/http"
	"strings"

	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/i18n"
	"github.com/gin-gonic/gin"
)

func CSRFTokenGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !requiresCSRFTokenCheck(c) {
			c.Next()
			return
		}

		token := strings.TrimSpace(c.GetHeader(constant.CSRFHeaderName))
		if !global.SESSION.CheckCSRFToken(c, token) {
			c.AbortWithStatusJSON(http.StatusForbidden, dto.Response{
				Code:    http.StatusForbidden,
				Message: i18n.GetMsgWithMap("ErrNotLogin", map[string]interface{}{"detail": "CSRF token invalid"}),
				Data:    nil,
			})
			return
		}

		c.Next()
	}
}

func requiresCSRFTokenCheck(c *gin.Context) bool {
	if c.GetBool("LOCAL_REQUEST") {
		return false
	}
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
	switch c.Request.URL.Path {
	case "/api/v2/core/auth/login",
		"/api/v2/core/auth/mfalogin",
		"/api/v2/core/auth/passkey/begin",
		"/api/v2/core/auth/passkey/finish",
		"/api/v2/core/auth/oidc/begin",
		"/api/v2/core/auth/oidc/finish":
		return false
	}
	if c.GetBool("API_AUTH") {
		return false
	}
	sessionID, err := c.Cookie(constant.SessionName)
	return err == nil && sessionID != ""
}
