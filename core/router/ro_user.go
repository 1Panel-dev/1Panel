package router

import (
	v2 "github.com/1Panel-dev/1Panel/core/app/api/v2"
	"github.com/1Panel-dev/1Panel/core/middleware"
	"github.com/gin-gonic/gin"
)

type UserRouter struct{}

func (s *UserRouter) InitRouter(Router *gin.RouterGroup) {
	userRouter := Router.Group("users").
		Use(middleware.SessionAuth()).
		Use(middleware.UserAuthMiddleware())
	{
		userApi := v2.ApiGroupApp.UserApi
		userRouter.GET("", userApi.ListUsers)
		userRouter.POST("", middleware.PermissionAuth("user:create"), userApi.CreateUser)
		userRouter.GET("/:id", userApi.GetUser)
		userRouter.PUT("", middleware.PermissionAuth("user:update"), userApi.UpdateUser)
		userRouter.DELETE("/:id", middleware.PermissionAuth("user:delete"), userApi.DeleteUser)

		// Password management
		userRouter.POST("/password/change", userApi.ChangePassword)
		userRouter.POST("/password/reset", middleware.PermissionAuth("user:password"), userApi.ResetPassword)

		// Permission management
		userRouter.GET("/:id/permissions", userApi.GetUserPermissions)
		userRouter.POST("/permissions", middleware.PermissionAuth("user:manage"), userApi.AssignPermissions)

		// Profile
		userRouter.GET("/profile", userApi.GetProfile)
		userRouter.GET("/:id/login-history", userApi.GetLoginHistory)
	}
}

type CommonRouter interface {
	InitRouter(Router *gin.RouterGroup)
}
