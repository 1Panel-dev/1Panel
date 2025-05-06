package middleware

import (
	"net/http"
	"strings"

	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/buserr"
	"github.com/gin-gonic/gin"
)

var whiteUrlList = map[string]struct{}{
	"/api/v2/dashboard/app/launcher/option": {},
	"/api/v2/websites/config":               {},
	"/api/v2/websites/waf/config":           {},
	"/api/v2/files/loadfile":                {},
	"/api/v2/files/size":                    {},
	"/api/v2/runtimes/sync":                 {},
	"/api/v2/toolbox/device/base":           {},

	"/api/v2/core/auth/login":     {},
	"/api/v2/core/logs/login":     {},
	"/api/v2/core/logs/operation": {},
	"/api/v2/core/auth/logout":    {},

	"/api/v2/apps/installed/loadport": {},
	"/api/v2/apps/installed/check":    {},
	"/api/v2/apps/installed/conninfo": {},
	"/api/v2/databases/load/file":     {},
	"/api/v2/databases/variables":     {},
	"/api/v2/databases/status":        {},
	"/api/v2/databases/baseinfo":      {},

	"/api/v2/xpack/waf/attack/stat":    {},
	"/api/v2/xpack/waf/config/website": {},
	"/api/v2/xpack/waf/relation/stat":  {},

	"/api/v2/xpack/monitor/stat":         {},
	"/api/v2/xpack/monitor/visitors":     {},
	"/api/v2/xpack/monitor/visitors/loc": {},
	"/api/v2/xpack/monitor/qps":          {},
	"/api/v2/xpack/monitor/logs/stat":    {},
	"/api/v2/xpack/monitor/websites":     {},
	"/api/v2/xpack/monitor/trend":        {},
	"/api/v2/xpack/monitor/rank":         {},
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
