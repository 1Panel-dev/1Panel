package middleware

import (
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/gin-gonic/gin"
)

func PasswordExpired() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, err := c.Cookie(constant.PasswordExpiredName)
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodePasswordExpired, constant.ErrTypePasswordExpired, nil)
			return
		}
		c.Next()
	}
}
