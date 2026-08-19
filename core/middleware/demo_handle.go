package middleware

import (
	"net/http"
	"strings"
	"sync"

	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/buserr"
	"github.com/gin-gonic/gin"
)

var (
	demoRouteMu sync.RWMutex

	demoReadOnlyPostRoutes = map[string]struct{}{
		"/api/v2/dashboard/app/launcher/option": {},
		"/api/v2/websites/config":               {},
		"/api/v2/websites/waf/config":           {},
		"/api/v2/websites/options":              {},
		"/api/v2/websites/rewrite":              {},
		"/api/v2/websites/dir":                  {},
		"/api/v2/websites/proxies":              {},
		"/api/v2/websites/auths":                {},
		"/api/v2/websites/leech":                {},
		"/api/v2/websites/redirect":             {},
		"/api/v2/files/size":                    {},
		"/api/v2/files/tree":                    {},
		"/api/v2/toolbox/device/base":           {},
		"/api/v2/files/user/group":              {},
		"/api/v2/files/mount":                   {},
		"/api/v2/hosts/ssh/log":                 {},
		"/api/v2/toolbox/clam/base":             {},
		"/api/v2/backups/record/size":           {},
		"/api/v2/containers/info":               {},
		"/api/v2/containers/list":               {},
		"/api/v2/containers/list/byimage":       {},
		"/api/v2/containers/users":              {},
		"/api/v2/containers/files/size":         {},
		"/api/v2/logs/system/read":              {},
		"/api/v2/logs/tasks/read":               {},

		"/api/v2/core/auth/login":           {},
		"/api/v2/core/auth/mfalogin":        {},
		"/api/v2/core/auth/passkey/begin":   {},
		"/api/v2/core/auth/passkey/finish":  {},
		"/api/v2/core/logs/login":           {},
		"/api/v2/core/logs/operation":       {},
		"/api/v2/core/auth/logout":          {},
		"/api/v2/core/settings/search/base": {},

		"/api/v2/apps/installed/loadport":         {},
		"/api/v2/apps/installed/check":            {},
		"/api/v2/apps/installed/conninfo":         {},
		"/api/v2/databases/common/info":           {},
		"/api/v2/databases/common/load/file":      {},
		"/api/v2/databases/load/file":             {},
		"/api/v2/databases/variables":             {},
		"/api/v2/databases/status":                {},
		"/api/v2/databases/baseinfo":              {},
		"/api/v2/backups/search/files":            {},
		"/api/v2/backups/record/search/bycronjob": {},
		"/api/v2/cronjobs/search/records":         {},

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
		"/api/v2/xpack/waf/cdn":              {},
		"/api/v2/xpack/tampers/search/log":   {},
		"/api/v2/xpack/tampers/search/file":  {},

		"/api/v2/core/nodes/list":                      {},
		"/api/v2/core/xpack/nodes/search/upgrade/logs": {},
	}

	demoBlockedGetRoutes = map[string]struct{}{
		"/api/v2/containers/exec":          {},
		"/api/v2/core/script/run":          {},
		"/api/v2/files/download":           {},
		"/api/v2/hosts/terminal/local":     {},
		"/api/v2/hosts/terminal/ssh":       {},
		"/api/v2/hosts/terminal/container": {},
		"/api/v2/process/:pid":             {},
	}
)

const demoReadOnlyContextKey = "DEMO_READ_ONLY_REQUEST"

func RegisterDemoReadOnlyPostRoutes(paths ...string) {
	demoRouteMu.Lock()
	defer demoRouteMu.Unlock()
	for _, routePath := range paths {
		if routePath != "" {
			demoReadOnlyPostRoutes[routePath] = struct{}{}
		}
	}
}

func RegisterDemoBlockedGetRoutes(paths ...string) {
	demoRouteMu.Lock()
	defer demoRouteMu.Unlock()
	for _, routePath := range paths {
		if routePath != "" {
			demoBlockedGetRoutes[routePath] = struct{}{}
		}
	}
}

func demoRequestPath(c *gin.Context) string {
	if routePath := c.FullPath(); routePath != "" {
		return routePath
	}
	return c.Request.URL.Path
}

func isDemoSearchRoute(routePath string) bool {
	return strings.HasSuffix(routePath, "/search") || strings.HasSuffix(routePath, "/ai-search")
}

func isDemoReadOnlyPost(routePath string) bool {
	if isDemoSearchRoute(routePath) {
		return true
	}
	demoRouteMu.RLock()
	defer demoRouteMu.RUnlock()
	for pattern := range demoReadOnlyPostRoutes {
		if matchDemoRoute(pattern, routePath) {
			return true
		}
	}
	return false
}

func isDemoBlockedGet(routePath string) bool {
	demoRouteMu.RLock()
	defer demoRouteMu.RUnlock()
	for pattern := range demoBlockedGetRoutes {
		if matchDemoRoute(pattern, routePath) {
			return true
		}
	}
	return false
}

func matchDemoRoute(pattern, routePath string) bool {
	if pattern == routePath {
		return true
	}
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	pathParts := strings.Split(strings.Trim(routePath, "/"), "/")
	if len(patternParts) != len(pathParts) {
		return false
	}
	for i := range patternParts {
		if strings.HasPrefix(patternParts[i], ":") {
			continue
		}
		if patternParts[i] != pathParts[i] {
			return false
		}
	}
	return true
}

func DemoHandle() gin.HandlerFunc {
	return func(c *gin.Context) {
		routePath := demoRequestPath(c)
		if c.Request.Method == http.MethodGet && !isDemoBlockedGet(routePath) {
			c.Set(demoReadOnlyContextKey, true)
			c.Next()
			return
		}
		if c.Request.Method == http.MethodPost && isDemoReadOnlyPost(routePath) {
			c.Set(demoReadOnlyContextKey, true)
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
