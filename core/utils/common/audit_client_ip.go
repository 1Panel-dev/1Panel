package common

import "github.com/gin-gonic/gin"

// AuditClientIPContextKey stores the global trusted-proxy result for one request.
// Only the server's early audit middleware should populate it.
const AuditClientIPContextKey = "1panel.audit_client_ip"

// AuditClientIP preserves a successfully authenticated API key's own proxy
// policy, then uses the global request snapshot, or finally the TCP peer.
// It deliberately does not use Gin's independently configured ClientIP parser.
func AuditClientIP(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return ""
	}
	if c.GetBool("API_AUTH") {
		if clientIP := c.GetString("API_AUTH_CLIENT_IP"); clientIP != "" {
			return clientIP
		}
	}
	if clientIP := c.GetString(AuditClientIPContextKey); clientIP != "" {
		return clientIP
	}
	return GetRealClientIP(c)
}
