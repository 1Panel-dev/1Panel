package middleware

import (
	"errors"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/gin-gonic/gin"
)

func WhiteAllow() gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(global.CONF.System.AllowIPs) == 0 {
			c.Next()
			return
		}
		clientIP := c.ClientIP()
		for _, ip := range strings.Split(global.CONF.System.AllowIPs, ",") {
			if len(ip) != 0 && ip == clientIP {
				c.Next()
				return
			}
		}
		helper.ErrorWithDetail(c, constant.CodeErrIP, constant.ErrTypeInternalServer, errors.New("IP address not allowed"))
	}
}
