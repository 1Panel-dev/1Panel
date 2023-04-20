package middleware

import (
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/constant"
	jwtUtils "github.com/1Panel-dev/1Panel/backend/utils/jwt"

	"github.com/gin-gonic/gin"
)

func JwtAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get(constant.JWTHeaderName)
		if token == "" {
			c.Next()
			return
		}
		j := jwtUtils.NewJWT()
		claims, err := j.ParseToken(token)
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrUnauthorized, constant.ErrTypeInternalServer, err)
			return
		}
		c.Set("claims", claims)
		c.Set("authMethod", constant.AuthMethodJWT)
		c.Next()
	}
}
