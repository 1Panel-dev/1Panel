package middleware

import (
	"encoding/base64"
	"github.com/1Panel-dev/1Panel/backend/app/repo"
	"github.com/gin-gonic/gin"
)

func SetPasswordPublicKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookieKey, _ := c.Cookie("panel_public_key")
		settingRepo := repo.NewISettingRepo()
		key, _ := settingRepo.Get(settingRepo.WithByKey("PASSWORD_PUBLIC_KEY"))
		base64Key := base64.StdEncoding.EncodeToString([]byte(key.Value))
		if base64Key == cookieKey {
			c.Next()
			return
		}
		c.SetCookie("panel_public_key", base64Key, 7*24*60*60, "/", "", false, false)
		c.Next()
	}
}
