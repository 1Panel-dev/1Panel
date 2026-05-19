package middleware

import (
	"fmt"
	"net/http"

	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/utils/common"
	"github.com/gin-gonic/gin"
)

const (
	localRequestContextKey = "LOCAL_REQUEST"
	localSSLReloadPath     = "/api/v2/core/settings/ssl/reload"
)

func LocalReqCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := common.GetRealClientIP(c)
		if !isLocalClientIP(clientIP) {
			abortLocalRequest(c, fmt.Sprintf("invalid client ip: %s", clientIP))
			return
		}
		if !common.ValidateLocalToken(c.GetHeader(common.LocalTokenHeader)) {
			abortLocalRequest(c, "local token invalid")
			return
		}
		c.Next()
	}
}

func IsLocalRequest(c *gin.Context) bool {
	if c.GetBool(localRequestContextKey) {
		return true
	}
	if c.Request.URL.Path != localSSLReloadPath {
		return false
	}
	if !isLocalClientIP(common.GetRealClientIP(c)) {
		return false
	}
	if !common.ValidateLocalToken(c.GetHeader(common.LocalTokenHeader)) {
		return false
	}
	c.Set(localRequestContextKey, true)
	return true
}

func isLocalClientIP(clientIP string) bool {
	return clientIP == "127.0.0.1" || clientIP == "::1"
}

func abortLocalRequest(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusForbidden, dto.Response{
		Code:    http.StatusForbidden,
		Message: message,
	})
}
