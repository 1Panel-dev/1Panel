package middleware

import (
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

var whiteUrlList = map[string]struct{}{
	"/api/v1/auth/login":          {},
	"/api/v1/websites/config":     {},
	"/api/v1/websites/waf/config": {},
	"/api/v1/files/loadfile":      {},
	"/api/v1/files/size":          {},
	"/api/v1/logs/operation":      {},
	"/api/v1/logs/login":          {},
	"/api/v1/auth/logout":         {},
}

func DemoHandle() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.Contains(c.Request.URL.Path, "search") || c.Request.Method == http.MethodGet {
			c.Next()
			return
		}
		if _, ok := whiteUrlList[c.Request.URL.Path]; ok {
			c.Next()
			return
		}

		c.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: buserr.New(constant.ErrDemoEnvironment).Error(),
		})
		c.Abort()
	}
}
