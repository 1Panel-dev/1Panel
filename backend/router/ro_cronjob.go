package router

import (
	v1 "github.com/1Panel-dev/1Panel/app/api/v1"
	"github.com/1Panel-dev/1Panel/middleware"

	"github.com/gin-gonic/gin"
)

type CronjobRouter struct{}

func (s *CronjobRouter) InitCronjobRouter(Router *gin.RouterGroup) {
	cmdRouter := Router.Group("cronjobs").Use(middleware.JwtAuth()).Use(middleware.SessionAuth())
	withRecordRouter := Router.Group("cronjobs").Use(middleware.JwtAuth()).Use(middleware.SessionAuth()).Use(middleware.OperationRecord())
	baseApi := v1.ApiGroupApp.BaseApi
	{
		withRecordRouter.POST("", baseApi.CreateCronjob)
		withRecordRouter.POST("/del", baseApi.DeleteCronjob)
		withRecordRouter.PUT(":id", baseApi.UpdateCronjob)
		withRecordRouter.POST("/status", baseApi.UpdateCronjobStatus)
		withRecordRouter.POST("/download", baseApi.TargetDownload)
		withRecordRouter.POST("/handle/:id", baseApi.HandleOnce)
		cmdRouter.POST("/search", baseApi.SearchCronjob)
		cmdRouter.POST("/search/records", baseApi.SearchJobRecords)
		cmdRouter.POST("/search/detail", baseApi.LoadRecordDetail)
	}
}
