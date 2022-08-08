package middleware

import (
	"time"

	"github.com/1Panel-dev/1Panel/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/constant"
	"github.com/1Panel-dev/1Panel/global"
	jwtUtils "github.com/1Panel-dev/1Panel/utils/jwt"
	"github.com/golang-jwt/jwt/v4"

	"github.com/gin-gonic/gin"
)

func JwtAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("authMethod", "")
		token := c.Request.Header.Get(global.CONF.JWT.HeaderName)
		if token == "" {
			c.Next()
			return
		}
		j := jwtUtils.NewJWT()
		claims, err := j.ParseToken(token)
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrUnauthorized, constant.ErrTypeToken, err)
			return
		}
		if claims.ExpiresAt.Unix()-time.Now().Unix() < claims.BufferTime {
			claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Second * time.Duration(global.CONF.JWT.ExpiresTime)))
		}
		c.Set("claims", claims)
		c.Set("authMethod", constant.AuthMethodJWT)
		c.Next()
	}
}
