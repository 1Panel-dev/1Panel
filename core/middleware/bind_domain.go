package middleware

import (
	"errors"
	"strings"

	"github.com/1Panel-dev/1Panel/core/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/gin-gonic/gin"
)

func BindDomain() gin.HandlerFunc {
	return func(c *gin.Context) {
		settingRepo := repo.NewISettingRepo()
		status, err := settingRepo.Get(repo.WithByKey("BindDomain"))
		if err != nil {
			helper.InternalServer(c, err)
			return
		}
		if len(status.Value) == 0 {
			c.Next()
			return
		}
		domains := c.Request.Host
		parts := strings.Split(c.Request.Host, ":")
		if len(parts) > 0 {
			domains = parts[0]
		}

		if domains != status.Value {
			if LoadErrCode() != 200 {
				helper.ErrResponse(c, LoadErrCode())
				return
			}
			helper.ErrorWithDetail(c, 311, "ErrInternalServer", errors.New("domain not allowed"))
			return
		}
		c.Next()
	}
}
