package router

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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
	RegisterImages(rootRouter)
	setStaticResource(rootRouter)
	rootRouter.GET("/assets/*filepath", func(c *gin.Context) {
		c.Writer.Header().Set("Cache-Control", fmt.Sprintf("private, max-age=%d", 2628000))
		if c.Request.URL.Path[len(c.Request.URL.Path)-1] == '/' {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
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

func RegisterImages(rootRouter *gin.RouterGroup) {
	staticDir := filepath.Join(global.CONF.Base.InstallDir, "1panel/uploads/theme")
	rootRouter.GET("/api/v2/images/*filename", func(c *gin.Context) {
		fileName := filepath.Base(c.Param("filename"))
		filePath := filepath.Join(staticDir, fileName)
		if !strings.HasPrefix(filePath, staticDir) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		f, err := os.Open(filePath)
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		defer f.Close()
		buf := make([]byte, 512)
		n, _ := f.Read(buf)
		content := buf[:n]
		mimeType := http.DetectContentType(buf[:n])
		if strings.Contains(string(content), "<svg") {
			mimeType = "image/svg+xml"
		}
		_, _ = f.Seek(0, io.SeekStart)
		c.Header("Content-Type", mimeType)
		_, _ = io.Copy(c.Writer, f)
	})
}

func setStaticResource(rootRouter *gin.RouterGroup) {
	rootRouter.GET("/api/v2/static/*filename", func(c *gin.Context) {
		c.Writer.Header().Set("Cache-Control", fmt.Sprintf("private, max-age=%d", 2628000))
		filename := c.Param("filename")
		filePath := "static" + filename
		data, err := web.Static.ReadFile(filePath)
		if err != nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		c.Writer.Write(data)
	})
}
