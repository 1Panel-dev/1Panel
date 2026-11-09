package providers

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

var opsReportRecordIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,127}$`)

const (
	OpsReportViewPermission   = "xpack_node_ops_report_view"
	OpsReportManagePermission = "xpack_node_ops_report_manage"
	OpsReportLegacyHeader     = "X-Panel-Ops-Report-Legacy-Read"
	OpsReportAPIPrefix        = "/api/v2/xpack/ops-report"
	OpsReportInternalPrefix   = "/api/v2/internal/ops-report/"
)

type OpsReportAccess struct{ AllowLegacy bool }

// Optional extension: community providers deliberately do not implement this.
type OpsReportAccessProvider interface {
	AuthorizeOpsReport(c *gin.Context, nodeName string, manage bool) (OpsReportAccess, error)
}

func IsOpsReportAPI(path string) bool {
	return strings.HasPrefix(path, OpsReportAPIPrefix+"/")
}

// Identify existing-record reads separately from operations which create new
// reports. Both kinds require an active license before they are proxied.
func IsOpsReportHistoryRead(method, path string) bool {
	if !strings.HasPrefix(path, OpsReportAPIPrefix+"/") {
		return false
	}
	if method == http.MethodPost && path == OpsReportAPIPrefix+"/runs/search" {
		return true
	}
	if method != http.MethodGet {
		return false
	}
	parts := strings.Split(strings.TrimPrefix(path, OpsReportAPIPrefix+"/"), "/")
	if len(parts) == 2 && opsReportRecordIDPattern.MatchString(parts[1]) && (parts[0] == "runs" || parts[0] == "requests") {
		return true
	}
	return len(parts) == 3 && parts[0] == "runs" && opsReportRecordIDPattern.MatchString(parts[1]) && parts[2] == "artifact"
}

// Core's pre-Agent archive contains sources from multiple former nodes and
// therefore requires its own administrator-only authorization and active license.
type OpsReportCoreArchiveAccessProvider interface {
	AuthorizeOpsReportCoreArchive(c *gin.Context) error
}

func IsOpsReportMaintenance(method, path string) bool {
	if method == http.MethodGet && (strings.HasPrefix(path, "/api/v2/core/ops-report/legacy/records/") || strings.HasPrefix(path, "/api/v2/core/ops-report/legacy/artifacts/")) {
		prefix := "/api/v2/core/ops-report/legacy/records/"
		if strings.HasPrefix(path, "/api/v2/core/ops-report/legacy/artifacts/") {
			prefix = "/api/v2/core/ops-report/legacy/artifacts/"
		}
		return opsReportRecordIDPattern.MatchString(strings.TrimPrefix(path, prefix))
	}
	if method != http.MethodPost {
		return false
	}
	switch path {
	case "/api/v2/core/ops-report/context", "/api/v2/core/ops-report/legacy/context", "/api/v2/core/ops-report/legacy/search":
		return true
	}
	return false
}
