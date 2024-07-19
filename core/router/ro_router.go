package router

import "github.com/gin-gonic/gin"

type CommonRouter interface {
	InitRouter(Router *gin.RouterGroup)
}
