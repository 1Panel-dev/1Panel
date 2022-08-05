package middleware

import (
	"errors"
	"time"

	"github.com/1Panel-dev/1Panel/app/dto"
	"github.com/1Panel-dev/1Panel/constant/errres"
	"github.com/1Panel-dev/1Panel/utils"

	"github.com/gin-gonic/gin"
)

func JwtAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("Authorization")
		re := dto.NewResult(c)
		if token == "" {
			re.Error(errres.JwtNotFound)
			return
		}
		j := utils.NewJWT()
		claims, err := j.ParseToken(token)
		if err != nil {
			if errors.Is(err, errres.TokenExpired) {
				re.Error(errres.JwtExpired)
				return
			}
			re.ErrorCode(errres.InvalidCommon, err.Error())
			return
		}
		if claims.ExpiresAt.Unix()-time.Now().Unix() < claims.BufferTime {
			//TODO 续签
		}
		c.Set("claims", claims)
		c.Next()
	}
}
