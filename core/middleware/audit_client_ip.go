package middleware

import (
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/utils/common"
	"github.com/gin-gonic/gin"
)

// AuditClientIP snapshots the global audit policy before any route middleware.
// Read it for every request so updates take effect without mutating Gin's engine
// proxy configuration while requests are running. Missing/invalid policy fails
// back to the TCP peer through the existing resolver.
func AuditClientIP() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := common.GetRealClientIP(c)
		settingRepo := repo.NewISettingRepo()
		if setting, err := settingRepo.Get(repo.WithByKey("AllowIPTrustedProxies")); err == nil {
			clientIP = common.ResolveClientIP(c, setting.Value)
		}
		c.Set(common.AuditClientIPContextKey, clientIP)
		c.Next()
	}
}
