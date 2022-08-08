package middleware

import (
	"github.com/1Panel-dev/1Panel/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/constant"
	"github.com/1Panel-dev/1Panel/global"
	"github.com/gin-gonic/gin"
)

func SessionAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if method, exist := c.Get("authMethod"); exist && method == constant.AuthMethodJWT {
			c.Next()
		}
		sID, err := c.Cookie(global.CONF.Session.SessionName)
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrUnauthorized, constant.ErrTypeToken, nil)
			return
		}
		sess, err := global.SESSION.Get(c.Request, sID)
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrUnauthorized, constant.ErrTypeToken, nil)
			return
		}
		if _, ok := sess.Values[global.CONF.Session.SessionUserKey]; !ok {
			helper.ErrorWithDetail(c, constant.CodeErrUnauthorized, constant.ErrTypeToken, nil)
			return
		}
		c.Next()
	}
}
