package middleware

import (
	"net/http"
	"strings"

	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/buserr"
	"github.com/gin-gonic/gin"
)

var whiteUrlList = map[string]struct{}{
	"/core/api/v2/auth/login":     {},
	"/api/v2/websites/config":     {},
	"/api/v2/websites/waf/config": {},
	"/api/v2/files/loadfile":      {},
	"/api/v2/files/size":          {},
	"/api/v2/logs/operation":      {},
	"/api/v2/logs/login":          {},
	"/core/api/v2/auth/logout":    {},

	"/api/v2/apps/installed/loadport": {},
	"/api/v2/apps/installed/check":    {},
	"/api/v2/apps/installed/conninfo": {},
	"/api/v2/databases/load/file":     {},
	"/api/v2/databases/variables":     {},
	"/api/v2/databases/status":        {},
	"/api/v2/databases/baseinfo":      {},

	"/api/v2/waf/attack/stat":    {},
	"/api/v2/waf/config/website": {},

	"/api/v2/monitor/stat":         {},
	"/api/v2/monitor/visitors":     {},
	"/api/v2/monitor/visitors/loc": {},
	"/api/v2/monitor/qps":          {},
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
			Message: buserr.New("ErrDemoEnvironment").Error(),
		})
		c.Abort()
	}
}
