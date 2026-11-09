//go:build !xpack && !enterprise

package router

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func opsReportProxyPolicy() agentProxyPolicy {
	return agentProxyPolicy{grant: func(c *gin.Context, _ bool) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/v2/xpack/ops-report/") {
			c.AbortWithStatus(http.StatusForbidden)
		}
	}}
}
