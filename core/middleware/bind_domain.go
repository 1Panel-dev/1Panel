package middleware

import (
	"errors"
	"strings"

	"github.com/1Panel-dev/1Panel/core/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/gin-gonic/gin"
)

func BindDomain() gin.HandlerFunc {
	return func(c *gin.Context) {
		settingRepo := repo.NewISettingRepo()
		commonRepo := repo.NewICommonRepo()
		status, err := settingRepo.Get(commonRepo.WithByKey("BindDomain"))
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
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
			if LoadErrCode("err-domain") != 200 {
				helper.ErrResponse(c, LoadErrCode("err-domain"))
				return
			}
			helper.ErrorWithDetail(c, constant.CodeErrDomain, constant.ErrTypeInternalServer, errors.New("domain not allowed"))
			return
		}
		c.Next()
	}
}
