package router

import (
	"1Panel/middlerware"
	"html/template"

	ginI18n "github.com/gin-contrib/i18n"
	"github.com/gin-gonic/gin"
)

func Routers() *gin.Engine {
	r := gin.Default()
	r.Use(middlerware.GinI18nLocalize())
	r.SetFuncMap(template.FuncMap{
		"Localize": ginI18n.GetMessage,
	})
	r.Use(middlerware.JwtAuth())
	return r
}
