package router

import (
	v1 "github.com/1Panel-dev/1Panel/backend/app/api/v1"
	"github.com/1Panel-dev/1Panel/backend/middleware"
	"github.com/gin-gonic/gin"
)

type RuntimeRouter struct {
}

func (r *RuntimeRouter) InitRuntimeRouter(Router *gin.RouterGroup) {
	groupRouter := Router.Group("runtimes")
	groupRouter.Use(middleware.JwtAuth()).Use(middleware.SessionAuth()).Use(middleware.PasswordExpired())

	baseApi := v1.ApiGroupApp.BaseApi
	{
		groupRouter.POST("/search", baseApi.SearchRuntimes)
		groupRouter.POST("", baseApi.CreateRuntime)
		groupRouter.POST("/del", baseApi.DeleteRuntime)
		groupRouter.POST("/update", baseApi.UpdateRuntime)
		groupRouter.GET("/:id", baseApi.GetRuntime)
	}
}
