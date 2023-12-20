package router

import (
	v1 "github.com/1Panel-dev/1Panel/backend/app/api/v1"
	"github.com/1Panel-dev/1Panel/backend/middleware"

	"github.com/gin-gonic/gin"
)

type CronjobRouter struct{}

func (s *CronjobRouter) InitRouter(Router *gin.RouterGroup) {
	cmdRouter := Router.Group("cronjobs").
		Use(middleware.JwtAuth()).
		Use(middleware.SessionAuth()).
		Use(middleware.PasswordExpired())
	baseApi := v1.ApiGroupApp.BaseApi
	{
		cmdRouter.POST("", baseApi.CreateCronjob)
		cmdRouter.POST("/del", baseApi.DeleteCronjob)
		cmdRouter.POST("/update", baseApi.UpdateCronjob)
		cmdRouter.POST("/status", baseApi.UpdateCronjobStatus)
		cmdRouter.POST("/handle", baseApi.HandleOnce)
		cmdRouter.POST("/download", baseApi.TargetDownload)
		cmdRouter.POST("/search", baseApi.SearchCronjob)
		cmdRouter.POST("/search/records", baseApi.SearchJobRecords)
		cmdRouter.POST("/records/log", baseApi.LoadRecordLog)
		cmdRouter.POST("/records/clean", baseApi.CleanRecord)
	}
}
