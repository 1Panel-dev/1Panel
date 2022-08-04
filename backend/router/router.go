package router

import (
	"github.com/1Panel-dev/1Panel/middlerware"
	ginI18n "github.com/gin-contrib/i18n"
	"github.com/gin-gonic/gin"
	"html/template"
)

func Routers() *gin.Engine {
	r := gin.Default()
	r.Use(middlerware.GinI18nLocalize())
	r.SetFuncMap(template.FuncMap{
		"Localize": ginI18n.GetMessage,
	})
	return r
}
