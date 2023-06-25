package middleware

import (
	"errors"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/repo"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/gin-gonic/gin"
)

func WhiteAllow() gin.HandlerFunc {
	return func(c *gin.Context) {
		settingRepo := repo.NewISettingRepo()
		status, err := settingRepo.Get(settingRepo.WithByKey("AllowIPs"))
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
			return
		}

		if len(status.Value) == 0 {
			c.Next()
			return
		}
		clientIP := c.ClientIP()
		for _, ip := range strings.Split(status.Value, ",") {
			if len(ip) != 0 && ip == clientIP {
				c.Next()
				return
			}
		}
		helper.ErrorWithDetail(c, constant.CodeErrIP, constant.ErrTypeInternalServer, errors.New("IP address not allowed"))
	}
}
