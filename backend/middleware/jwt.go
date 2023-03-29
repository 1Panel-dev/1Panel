package middleware

import (
	"strconv"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/repo"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	jwtUtils "github.com/1Panel-dev/1Panel/backend/utils/jwt"
	"github.com/golang-jwt/jwt/v4"

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
		if claims.ExpiresAt.Unix()-time.Now().Unix() < claims.BufferTime {
			settingRepo := repo.NewISettingRepo()
			setting, err := settingRepo.Get(settingRepo.WithByKey("SessionTimeout"))
			if err != nil {
				global.LOG.Errorf("create operation record failed, err: %v", err)
			}
			lifeTime, _ := strconv.Atoi(setting.Value)
			claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Second * time.Duration(lifeTime)))
		}
		c.Set("claims", claims)
		c.Set("authMethod", constant.AuthMethodJWT)
		c.Next()
	}
}
