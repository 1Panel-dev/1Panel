package middleware

import (
	"github.com/1Panel-dev/1Panel/app/model"
	"github.com/1Panel-dev/1Panel/global"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		operateLog := model.OperateLog{
			Path:      path,
			IP:        c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
		}
		global.DB.Model(model.OperateLog{}).Save(&operateLog)
		c.Next()
	}
}

//TODO 根据URL写操作日志
