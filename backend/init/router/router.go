package router

import (
	"html/template"
	"net/http"

	v1 "github.com/1Panel-dev/1Panel/backend/app/api/v1"
	"github.com/1Panel-dev/1Panel/backend/i18n"
	"github.com/1Panel-dev/1Panel/backend/middleware"
	rou "github.com/1Panel-dev/1Panel/backend/router"
	"github.com/1Panel-dev/1Panel/cmd/server/docs"
	"github.com/1Panel-dev/1Panel/cmd/server/web"
	ginI18n "github.com/gin-contrib/i18n"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func setWebStatic(rootRouter *gin.Engine) {
	rootRouter.StaticFS("/kubepi/login/onepanel", http.FS(web.IndexHtml))
	rootRouter.StaticFS("/favicon.ico", http.FS(web.Favicon))
	rootRouter.GET("/assets/*filepath", func(c *gin.Context) {
		staticServer := http.FileServer(http.FS(web.Assets))
		staticServer.ServeHTTP(c.Writer, c.Request)
	})

	rootRouter.GET("/", func(c *gin.Context) {
		staticServer := http.FileServer(http.FS(web.IndexHtml))
		staticServer.ServeHTTP(c.Writer, c.Request)
	})
	rootRouter.NoRoute(func(c *gin.Context) {
		c.Writer.WriteHeader(http.StatusOK)
		_, _ = c.Writer.Write(web.IndexByte)
		c.Writer.Header().Add("Accept", "text/html")
		c.Writer.Flush()
	})
}

func Routers() *gin.Engine {
	Router := gin.Default()

	docs.SwaggerInfo.BasePath = "/api/v1"
	Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	Router.Use(middleware.OperationLog())
	Router.Use(middleware.CSRF())
	Router.Use(middleware.LoadCsrfToken())

	setWebStatic(Router)

	Router.Use(i18n.GinI18nLocalize())
	Router.GET("/api/v1/info", v1.ApiGroupApp.BaseApi.GetSafetyStatus)
	Router.GET("/api/v1/:code", v1.ApiGroupApp.BaseApi.SafeEntrance)

	Router.SetFuncMap(template.FuncMap{
		"Localize": ginI18n.GetMessage,
	})
	Router.Use(middleware.JwtAuth())

	systemRouter := rou.RouterGroupApp

	PublicGroup := Router.Group("")
	{
		PublicGroup.GET("/health", func(c *gin.Context) {
			c.JSON(200, "ok")
		})
	}

	PrivateGroup := Router.Group("/api/v1")
	//PrivateGroup.Use(middleware.SafetyAuth())
	{
		systemRouter.InitBaseRouter(PrivateGroup)
		systemRouter.InitDashboardRouter(PrivateGroup)
		systemRouter.InitHostRouter(PrivateGroup)
		systemRouter.InitBackupRouter(PrivateGroup)
		systemRouter.InitGroupRouter(PrivateGroup)
		systemRouter.InitContainerRouter(PrivateGroup)
		systemRouter.InitCommandRouter(PrivateGroup)
		systemRouter.InitTerminalRouter(PrivateGroup)
		systemRouter.InitMonitorRouter(PrivateGroup)
		systemRouter.InitLogRouter(PrivateGroup)
		systemRouter.InitFileRouter(PrivateGroup)
		systemRouter.InitCronjobRouter(PrivateGroup)
		systemRouter.InitSettingRouter(PrivateGroup)
		systemRouter.InitAppRouter(PrivateGroup)
		systemRouter.InitWebsiteRouter(PrivateGroup)
		systemRouter.InitWebsiteGroupRouter(PrivateGroup)
		systemRouter.InitWebsiteDnsAccountRouter(PrivateGroup)
		systemRouter.InitDatabaseRouter(PrivateGroup)
		systemRouter.InitWebsiteSSLRouter(PrivateGroup)
		systemRouter.InitWebsiteAcmeAccountRouter(PrivateGroup)
		systemRouter.InitNginxRouter(PrivateGroup)
	}

	return Router
}
