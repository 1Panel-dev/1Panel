package middleware

import (
	"errors"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/gin-gonic/gin"
)

func BindDomain() gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(global.CONF.System.BindDomain) == 0 {
			c.Next()
			return
		}
		domains := c.Request.Host
		parts := strings.Split(c.Request.Host, ":")
		if len(parts) > 0 {
			domains = parts[0]
		}

		if domains != global.CONF.System.BindDomain {
			helper.ErrorWithDetail(c, constant.CodeErrDomain, constant.ErrTypeInternalServer, errors.New("domain not allowed"))
			return
		}
		c.Next()
	}
}
