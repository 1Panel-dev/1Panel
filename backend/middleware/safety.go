package middleware

import (
	"github.com/1Panel-dev/1Panel/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/app/service"
	"github.com/1Panel-dev/1Panel/constant"
	"github.com/gin-gonic/gin"
)

func SafetyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := service.NewIAuthService().SafetyStatus(c); err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrUnSafety, constant.ErrTypeNotSafety, nil)
			return
		}
		c.Next()
	}
}
