package middleware

import (
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/service"
	"github.com/1Panel-dev/1Panel/backend/constant"
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
