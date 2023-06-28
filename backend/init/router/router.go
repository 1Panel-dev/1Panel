package router

import (
	"github.com/gin-contrib/gzip"
	"html/template"
	"net/http"

	"github.com/1Panel-dev/1Panel/backend/global"
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

func setWebStatic(rootRouter *gin.RouterGroup) {
	rootRouter.StaticFS("/public", http.FS(web.Favicon))
	rootRouter.GET("/assets/*filepath", func(c *gin.Context) {
		staticServer := http.FileServer(http.FS(web.Assets))
		staticServer.ServeHTTP(c.Writer, c.Request)
	})
	rootRouter.GET("/", func(c *gin.Context) {
		staticServer := http.FileServer(http.FS(web.IndexHtml))
		staticServer.ServeHTTP(c.Writer, c.Request)
	})
}

func Routers() *gin.Engine {
	Router := gin.Default()
	Router.Use(middleware.OperationLog())
	// Router.Use(middleware.CSRF())
	// Router.Use(middleware.LoadCsrfToken())
	if global.CONF.System.IsDemo {
		Router.Use(middleware.DemoHandle())
	}

	Router.NoRoute(func(c *gin.Context) {
		c.Writer.WriteHeader(http.StatusOK)
		_, _ = c.Writer.Write(web.IndexByte)
		c.Writer.Header().Add("Accept", "text/html")
		c.Writer.Flush()
	})

	Router.Use(i18n.GinI18nLocalize())
	Router.SetFuncMap(template.FuncMap{
		"Localize": ginI18n.GetMessage,
	})

	systemRouter := rou.RouterGroupApp
	swaggerRouter := Router.Group("1panel")
	docs.SwaggerInfo.BasePath = "/api/v1"
	swaggerRouter.Use(middleware.JwtAuth()).Use(middleware.SessionAuth()).GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	PublicGroup := Router.Group("")
	{
		PublicGroup.GET("/health", func(c *gin.Context) {
			c.JSON(200, "ok")
		})
		PublicGroup.Use(gzip.Gzip(gzip.DefaultCompression))
		setWebStatic(PublicGroup)
	}
	PrivateGroup := Router.Group("/api/v1")
	PrivateGroup.Use(middleware.WhiteAllow())
	PrivateGroup.Use(middleware.BindDomain())
	PrivateGroup.Use(middleware.GlobalLoading())
	{
		systemRouter.InitBaseRouter(PrivateGroup)
		systemRouter.InitDashboardRouter(PrivateGroup)
		systemRouter.InitHostRouter(PrivateGroup)
		systemRouter.InitContainerRouter(PrivateGroup)
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
		systemRouter.InitRuntimeRouter(PrivateGroup)
		systemRouter.InitProcessRouter(PrivateGroup)
	}

	return Router
}
