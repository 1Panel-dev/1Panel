package router

import (
	"encoding/base64"
	"fmt"
	"github.com/1Panel-dev/1Panel/backend/app/service"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/cmd/server/res"
	"net/http"
	"regexp"
	"strconv"
	"strings"

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
)

var (
	Router *gin.Engine
)

func toIndexHtml(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Writer.WriteHeader(http.StatusOK)
	_, _ = c.Writer.Write(web.IndexByte)
	c.Writer.Flush()
}

func isEntrancePath(c *gin.Context) bool {
	entrance := service.NewIAuthService().GetSecurityEntrance()
	if entrance != "" && strings.TrimSuffix(c.Request.URL.Path, "/") == "/"+entrance {
		return true
	}
	return false
}

func isFrontendPath(c *gin.Context) bool {
	reqUri := strings.TrimSuffix(c.Request.URL.Path, "/")
	if _, ok := constant.WebUrlMap[reqUri]; ok {
		return true
	}
	for _, route := range constant.DynamicRoutes {
		if match, _ := regexp.MatchString(route, reqUri); match {
			return true
		}
	}
	return false
}

func checkFrontendPath(c *gin.Context) bool {
	if !isFrontendPath(c) {
		return false
	}
	authService := service.NewIAuthService()
	if authService.GetSecurityEntrance() != "" {
		return authService.IsLogin(c)
	}
	return true
}

func checkEntrance(c *gin.Context) bool {
	authService := service.NewIAuthService()
	entrance := authService.GetSecurityEntrance()
	if entrance == "" {
		return true
	}

	cookieValue, err := c.Cookie("SecurityEntrance")
	if err != nil {
		return false
	}
	entranceValue, err := base64.StdEncoding.DecodeString(cookieValue)
	if err != nil {
		return false
	}
	return string(entranceValue) == entrance
}

func handleNoRoute(c *gin.Context) {
	resPage, err := service.NewIAuthService().GetResponsePage()
	if err != nil {
		c.String(http.StatusInternalServerError, "Internal Server Error")
		return
	}
	if resPage == "444" {
		c.String(444, "")
		return
	}

	file := fmt.Sprintf("html/%s.html", resPage)
	data, err := res.ErrorMsg.ReadFile(file)
	if err != nil {
		c.String(http.StatusInternalServerError, "Internal Server Error")
		return
	}
	statusCode, err := strconv.Atoi(resPage)
	if err != nil {
		c.String(http.StatusInternalServerError, "Internal Server Error")
		return
	}
	c.Data(statusCode, "text/html; charset=utf-8", data)
}

func setWebStatic(rootRouter *gin.RouterGroup) {
	rootRouter.StaticFS("/public", http.FS(web.Favicon))
	rootRouter.StaticFS("/favicon.ico", http.FS(web.Favicon))
	rootRouter.Static("/api/v1/images", "./uploads")

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
			if currentEntrance == "" || currentEntrance != entrance {
				handleNoRoute(c)
				return
			}
			c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
			c.Writer.WriteHeader(http.StatusOK)
			_, _ = c.Writer.Write(web.IndexByte)
			c.Writer.Flush()
		})
	}
	rootRouter.GET("/", func(c *gin.Context) {
		if !checkEntrance(c) {
			handleNoRoute(c)
			return
		}
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

	Router.Use(middleware.WhiteAllow())
	Router.Use(middleware.BindDomain())
	Router.Use(middleware.SetPasswordPublicKey())

	Router.NoRoute(func(c *gin.Context) {
		if checkFrontendPath(c) {
			toIndexHtml(c)
			return
		}
		if isEntrancePath(c) {
			toIndexHtml(c)
			return
		}
		handleNoRoute(c)
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

	PrivateGroup.Use(middleware.GlobalLoading())
	for _, router := range rou.RouterGroupApp {
		router.InitRouter(PrivateGroup)
	}

	return Router
}
