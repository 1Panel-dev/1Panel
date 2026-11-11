//go:build xpack || enterprise

package router

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/core/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/core/utils/xpack"
	report "github.com/1Panel-dev/1Panel/core/xpack/utils/opsreport"
	"github.com/1Panel-dev/1Panel/core/xpack/utils/opsreport/contracts"
	"github.com/gin-gonic/gin"
)

func opsReportProxyPolicy() agentProxyPolicy {
	return agentProxyPolicy{
		clear: func(r *http.Request) {
			// Older Agents may still recognize this retired header during upgrades.
			r.Header.Del("X-Panel-Ops-Report-Legacy-Read")
			if contracts.IsOpsReportAPI(r.URL.Path) {
				r.Header.Del("X-Panel-User")
			}
		},
		grant: func(c *gin.Context, _ bool) {
			if !contracts.IsOpsReportAPI(c.Request.URL.Path) {
				return
			}
			authorizer, ok := xpack.AuthProvider.(contracts.OpsReportAccessProvider)
			if !ok {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			name := c.Query("operateNode")
			if name == "" || name == "undefined" {
				name = c.GetHeader("CurrentNode")
			}
			name, err := url.QueryUnescape(name)
			if err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
			if name == "" {
				name = "local"
			}
			manage := strings.HasSuffix(c.Request.URL.Path, "/generate") || strings.HasSuffix(c.Request.URL.Path, "/download") || strings.HasSuffix(c.Request.URL.Path, "/setting/update") || strings.HasSuffix(c.Request.URL.Path, "/runs/delete")
			if err := authorizer.AuthorizeOpsReport(c, name, manage); err != nil {
				helper.ErrorWithDetail(c, http.StatusForbidden, "ErrRBAC", err)
				c.Abort()
				return
			}
			node, err := report.LoadNode(name)
			if err != nil || node.ID == 0 {
				helper.ErrorWithDetail(c, http.StatusBadRequest, "ErrNoSuchNode", err)
				c.Abort()
				return
			}
			policy, err := report.Evaluate(node, time.Now())
			if err != nil || !policy.Active {
				helper.ErrorWithDetail(c, http.StatusForbidden, "ErrXpackNotFound", err)
				c.Abort()
				return
			}
			capability, err := report.Capabilities(c.Request.Context(), node)
			if err != nil {
				if errors.Is(err, report.ErrReportAgentUpgrade) {
					helper.BadRequest(c, err)
				} else {
					helper.ErrorWithDetail(c, http.StatusServiceUnavailable, "ErrInternalServer", err)
				}
				c.Abort()
				return
			}
			report.InitializeAgentSettings(c.Request.Context(), node, capability)
		},
	}
}
