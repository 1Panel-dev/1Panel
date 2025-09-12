package router

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"path"

	"github.com/1Panel-dev/1Panel/core/app/service"
	"github.com/1Panel-dev/1Panel/core/cmd/server/docs"
	"github.com/1Panel-dev/1Panel/core/cmd/server/web"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/i18n"
	"github.com/1Panel-dev/1Panel/core/middleware"
	rou "github.com/1Panel-dev/1Panel/core/router"
	"github.com/1Panel-dev/1Panel/core/utils/security"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var (
	Router *gin.Engine
)

func setWebStatic(rootRouter *gin.RouterGroup) {
	rootRouter.StaticFS("/public", http.FS(web.Favicon))
	rootRouter.StaticFS("/favicon.ico", http.FS(web.Favicon))
	rootRouter.Static("/api/v2/images", path.Join(global.CONF.Base.InstallDir, "1panel/uploads/theme"))
	rootRouter.GET("/assets/*filepath", func(c *gin.Context) {
		c.Writer.Header().Set("Cache-Control", fmt.Sprintf("private, max-age=%d", 3600))
		staticServer := http.FileServer(http.FS(web.Assets))
		staticServer.ServeHTTP(c.Writer, c.Request)
	})
	authService := service.NewIAuthService()
	entrance := authService.GetSecurityEntrance()
	if entrance != "" {
		rootRouter.GET("/"+entrance, func(c *gin.Context) {
			currentEntrance := authService.GetSecurityEntrance()
			if currentEntrance != entrance {
				security.HandleNotSecurity(c, "")
				return
			}
			security.ToIndexHtml(c)
		})
	}
	rootRouter.GET("/", func(c *gin.Context) {
		if !security.CheckSecurity(c) {
			return
		}
		entrance = authService.GetSecurityEntrance()
		if entrance != "" {
			entranceValue := base64.StdEncoding.EncodeToString([]byte(entrance))
			c.SetCookie("SecurityEntrance", entranceValue, 0, "", "", false, true)
		}
		staticServer := http.FileServer(http.FS(web.IndexHtml))
		staticServer.ServeHTTP(c.Writer, c.Request)
	})
}

func Routers() *gin.Engine {
	Router = gin.Default()
	Router.Use(i18n.UseI18n())
	Router.Use(middleware.WhiteAllow())
	Router.Use(middleware.BindDomain())

	swaggerRouter := Router.Group("1panel")
	docs.SwaggerInfo.BasePath = "/api/v2"
	swaggerRouter.Use(middleware.SessionAuth()).GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	PublicGroup := Router.Group("")
	{
		PublicGroup.Use(gzip.Gzip(gzip.DefaultCompression))
		setWebStatic(PublicGroup)
	}
	if global.CONF.Base.IsDemo {
		Router.Use(middleware.DemoHandle())
	}

	Router.Use(middleware.OperationLog())
	Router.Use(middleware.GlobalLoading())
	Router.Use(middleware.PasswordExpired())
	Router.Use(middleware.ApiAuth())

	PrivateGroup := Router.Group("/api/v2/core")
	PrivateGroup.Use(middleware.SetPasswordPublicKey())
	for _, router := range rou.RouterGroupApp {
		router.InitRouter(PrivateGroup)
	}

	Router.Use(Proxy())
	Router.NoRoute(func(c *gin.Context) {
		if !security.HandleNotRoute(c) {
			return
		}
		security.HandleNotSecurity(c, "")
	})

	return Router
}
