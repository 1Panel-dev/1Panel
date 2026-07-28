package router

import (
	v2 "github.com/1Panel-dev/1Panel/agent/app/api/v2"
	"github.com/gin-gonic/gin"
)

type WebsiteTemplateRouter struct {
}

func (a *WebsiteTemplateRouter) InitRouter(Router *gin.RouterGroup) {
	groupRouter := Router.Group("websites/templates")

	baseApi := v2.ApiGroupApp.BaseApi
	{
		groupRouter.POST("/search", baseApi.PageWebsiteTemplate)
		groupRouter.POST("", baseApi.CreateWebsiteTemplate)
		groupRouter.POST("/update", baseApi.UpdateWebsiteTemplate)
		groupRouter.POST("/del", baseApi.DeleteWebsiteTemplate)
		groupRouter.POST("/get", baseApi.GetWebsiteTemplate)
		groupRouter.POST("/upload", baseApi.UploadTemplateZip)
		groupRouter.POST("/preview", baseApi.PreviewWebsiteTemplate)
		groupRouter.POST("/outputs/search", baseApi.PageWebsiteTemplateOutput)
		groupRouter.POST("/outputs", baseApi.CreateWebsiteTemplateOutput)
		groupRouter.POST("/outputs/del", baseApi.DeleteWebsiteTemplateOutput)
		groupRouter.POST("/outputs/get", baseApi.GetWebsiteTemplateOutput)
		groupRouter.POST("/outputs/import", baseApi.ImportTemplateOutput)
	}
}
