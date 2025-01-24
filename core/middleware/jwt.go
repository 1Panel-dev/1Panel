package middleware

import (
	"fmt"
	"strings"

	"github.com/1Panel-dev/1Panel/core/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/core/constant"
	jwtUtils "github.com/1Panel-dev/1Panel/core/utils/jwt"

	"github.com/gin-gonic/gin"
)

func JwtAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/v2/core/auth") {
			c.Next()
			return
		}
		token := c.Request.Header.Get(constant.JWTHeaderName)
		if token == "" {
			c.Next()
			return
		}
		j := jwtUtils.NewJWT()
		claims, err := j.ParseToken(token)
		if err != nil {
			helper.BadAuth(c, "ErrInternalServer", err)
			return
		}
		if claims.BaseClaims.IsAgent {
			if strings.HasPrefix(c.Request.URL.Path, "/api/v2/agent/") {
				c.Set("claims", claims)
				c.Set("authMethod", constant.AuthMethodJWT)
				c.Next()
				return
			} else {
				helper.BadAuth(c, "ErrInternalServer", fmt.Errorf("err token from request"))
				return
			}
		}
		if strings.HasPrefix(c.Request.URL.Path, "/api/v2/agent/") {
			helper.BadAuth(c, "ErrInternalServer", fmt.Errorf("err token from request"))
			return
		}
		c.Set("claims", claims)
		c.Set("authMethod", constant.AuthMethodJWT)
		c.Next()
	}
}
