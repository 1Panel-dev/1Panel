package middleware

import "strings"

func ShouldProxyToAgent(reqPath string) bool {
	if strings.HasPrefix(reqPath, "/1panel/swagger") || !strings.HasPrefix(reqPath, "/api/v2") {
		return false
	}
	if strings.HasPrefix(reqPath, "/api/v2/core") && !strings.HasPrefix(reqPath, "/api/v2/core/xpack") {
		return false
	}
	return true
}
