package router

import (
	v1 "github.com/1Panel-dev/1Panel/backend/app/api/v1"
	"github.com/1Panel-dev/1Panel/backend/middleware"
	"github.com/gin-gonic/gin"
)

type FileRouter struct {
}

func (f *FileRouter) InitRouter(Router *gin.RouterGroup) {
	fileRouter := Router.Group("files")
	fileRouter.Use(middleware.JwtAuth()).Use(middleware.SessionAuth()).Use(middleware.PasswordExpired())
	baseApi := v1.ApiGroupApp.BaseApi
	{
		fileRouter.POST("/search", baseApi.ListFiles)
		fileRouter.POST("/upload/search", baseApi.SearchUploadWithPage)
		fileRouter.POST("/tree", baseApi.GetFileTree)
		fileRouter.POST("", baseApi.CreateFile)
		fileRouter.POST("/del", baseApi.DeleteFile)
		fileRouter.POST("/batch/del", baseApi.BatchDeleteFile)
		fileRouter.POST("/mode", baseApi.ChangeFileMode)
		fileRouter.POST("/owner", baseApi.ChangeFileOwner)
		fileRouter.POST("/compress", baseApi.CompressFile)
		fileRouter.POST("/decompress", baseApi.DeCompressFile)
		fileRouter.POST("/content", baseApi.GetContent)
		fileRouter.POST("/save", baseApi.SaveContent)
		fileRouter.POST("/check", baseApi.CheckFile)
		fileRouter.POST("/batch/check", baseApi.BatchCheckFiles)
		fileRouter.POST("/upload", baseApi.UploadFiles)
		fileRouter.POST("/chunkupload", baseApi.UploadChunkFiles)
		fileRouter.POST("/rename", baseApi.ChangeFileName)
		fileRouter.POST("/wget", baseApi.WgetFile)
		fileRouter.POST("/move", baseApi.MoveFile)
		fileRouter.GET("/download", baseApi.Download)
		fileRouter.POST("/chunkdownload", baseApi.DownloadChunkFiles)
		fileRouter.POST("/size", baseApi.Size)
		fileRouter.GET("/ws", baseApi.Ws)
		fileRouter.GET("/keys", baseApi.Keys)
		fileRouter.POST("/read", baseApi.ReadFileByLine)
		fileRouter.POST("/batch/role", baseApi.BatchChangeModeAndOwner)

		fileRouter.POST("/recycle/search", baseApi.SearchRecycleBinFile)
		fileRouter.POST("/recycle/reduce", baseApi.ReduceRecycleBinFile)
		fileRouter.POST("/recycle/clear", baseApi.ClearRecycleBinFile)
		fileRouter.GET("/recycle/status", baseApi.GetRecycleStatus)

		fileRouter.POST("/favorite/search", baseApi.SearchFavorite)
		fileRouter.POST("/favorite", baseApi.CreateFavorite)
		fileRouter.POST("/favorite/del", baseApi.DeleteFavorite)

	}
}
