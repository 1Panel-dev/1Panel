package router

import (
	v1 "github.com/1Panel-dev/1Panel/internal/api/v1"
	"github.com/gin-gonic/gin"
)

func Routers() *gin.Engine {
	r := gin.Default()

	r.GET("/api/v1/users", v1.GetUser)
	return r
}
