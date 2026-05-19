package middleware

import "strings"

// ShouldProxyToAgent reports whether a request will eventually be forwarded to
// an agent (local or remote) instead of being handled by core itself. This is
// the single source of truth for the router-level Proxy gate and the
// operation-log middleware that needs to decide whether to ask the agent to
// resolve operation metadata.
//
// Rules:
//   - Non-`/api/v2/...` (including swagger UI assets) is always handled by core.
//   - `/api/v2/core/...` is handled by core, except `/api/v2/core/xpack/...`
//     which is allowed to be re-routed to the active node.
//   - Everything else is forwarded to the agent.
func ShouldProxyToAgent(reqPath string) bool {
	if strings.HasPrefix(reqPath, "/1panel/swagger") || !strings.HasPrefix(reqPath, "/api/v2") {
		return false
	}
	if strings.HasPrefix(reqPath, "/api/v2/core") && !strings.HasPrefix(reqPath, "/api/v2/core/xpack") {
		return false
	}
	return true
}
