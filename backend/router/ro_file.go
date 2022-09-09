package router

import (
	v1 "github.com/1Panel-dev/1Panel/app/api/v1"
	"github.com/1Panel-dev/1Panel/middleware"
	"github.com/gin-gonic/gin"
)

type FileRouter struct {
}

func (f *FileRouter) InitFileRouter(Router *gin.RouterGroup) {
	fileRouter := Router.Group("files")
	fileRouter.Use(middleware.JwtAuth()).Use(middleware.SessionAuth())
	//withRecordRouter := fileRouter.Use(middleware.OperationRecord())
	baseApi := v1.ApiGroupApp.BaseApi
	{
		fileRouter.POST("/search", baseApi.ListFiles)
		fileRouter.POST("/tree", baseApi.GetFileTree)
		fileRouter.POST("", baseApi.CreateFile)
		fileRouter.POST("/del", baseApi.DeleteFile)
		fileRouter.POST("/mode", baseApi.ChangeFileMode)
		fileRouter.POST("/compress", baseApi.CompressFile)
		fileRouter.POST("/decompress", baseApi.DeCompressFile)
		fileRouter.POST("/content", baseApi.GetContent)
		fileRouter.POST("/save", baseApi.SaveContent)
		fileRouter.POST("/upload", baseApi.UploadFiles)
		fileRouter.POST("/rename", baseApi.ChangeFileName)
		fileRouter.POST("/wget", baseApi.WgetFile)
		fileRouter.POST("/move", baseApi.MoveFile)
		fileRouter.POST("/download", baseApi.Download)
		fileRouter.POST("/size", baseApi.Size)
	}

}
