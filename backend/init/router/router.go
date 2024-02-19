package router

import (
	"fmt"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/i18n"
	"github.com/1Panel-dev/1Panel/backend/middleware"
	rou "github.com/1Panel-dev/1Panel/backend/router"
	"github.com/1Panel-dev/1Panel/cmd/server/docs"
	"github.com/1Panel-dev/1Panel/cmd/server/web"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"net/http"
)

var (
	Router *gin.Engine
)

func setWebStatic(rootRouter *gin.RouterGroup) {
	rootRouter.StaticFS("/public", http.FS(web.Favicon))
	rootRouter.Use(func(c *gin.Context) {
		c.Next()
	})
	rootRouter.GET("/assets/*filepath", func(c *gin.Context) {
		c.Writer.Header().Set("Cache-Control", fmt.Sprintf("private, max-age=%d", 3600))
		staticServer := http.FileServer(http.FS(web.Assets))
		staticServer.ServeHTTP(c.Writer, c.Request)
	})
	rootRouter.GET("/", func(c *gin.Context) {
		staticServer := http.FileServer(http.FS(web.IndexHtml))
		staticServer.ServeHTTP(c.Writer, c.Request)
	})
}

func Routers() *gin.Engine {
	Router = gin.Default()
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

	Router.Use(i18n.UseI18n())

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
	for _, router := range rou.RouterGroupApp {
		router.InitRouter(PrivateGroup)
	}

	return Router
}
