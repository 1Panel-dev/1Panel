package router

import (
	"github.com/1Panel-dev/1Panel/docs"
	"github.com/1Panel-dev/1Panel/middlerware"
	ginI18n "github.com/gin-contrib/i18n"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"html/template"
)

func Routers() *gin.Engine {
	r := gin.Default()
	docs.SwaggerInfo.BasePath = "/api/v1"
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	r.Use(middlerware.GinI18nLocalize())
	r.SetFuncMap(template.FuncMap{
		"Localize": ginI18n.GetMessage,
	})
	r.Use(middlerware.JwtAuth())
	return r
}
