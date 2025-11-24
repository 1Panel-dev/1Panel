package swagger

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/1Panel-dev/1Panel/core/cmd/server/docs"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files/v2"
)

var matcher = regexp.MustCompile(fileMatchPattern)

const swaggerDocFile = "doc.json"

const fileMatchPattern = `(.*)(index\.html|index\.css|swagger-initializer\.js|doc\.json|favicon-16x16\.png|favicon-32x32\.png|/oauth2-redirect\.html|swagger-ui\.css|swagger-ui\.css\.map|swagger-ui\.js|swagger-ui\.js\.map|swagger-ui-bundle\.js|swagger-ui-bundle\.js\.map|swagger-ui-standalone-preset\.js|swagger-ui-standalone-preset\.js\.map)[?|.]*`
const CustomSwaggerInitializerJS = `window.onload = function() {
  //<editor-fold desc="Changeable Configuration Block">
  window.ui = SwaggerUIBundle({
    url: "{{.URL}}",
    dom_id: '#swagger-ui',
    deepLinking: true,
    presets: [
      SwaggerUIBundle.presets.apis,
      SwaggerUIStandalonePreset
    ],
    plugins: [
      SwaggerUIBundle.plugins.DownloadUrl
    ],
    layout: "StandaloneLayout"
  });

  //</editor-fold>
};
`

func SwaggerHandler() gin.HandlerFunc {
	fileServer := http.StripPrefix("/1panel/swagger/", http.FileServer(http.FS(swaggerfiles.FS)))

	return func(c *gin.Context) {
		path := c.Request.URL.Path
		path = strings.TrimPrefix(path, "/1panel/swagger")
		if !matcher.MatchString(path) && path != "/" {
			// 404
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		switch path {
		case "/doc.json":
			c.Header("Content-Type", "application/json; charset=utf-8")
			c.Header("Cache-Control", "private, max-age=600")
			c.String(http.StatusOK, docs.SwaggerInfo.ReadDoc())
		case "/swagger-initializer.js":
			c.Header("Content-Type", "application/javascript; charset=utf-8")
			c.Header("Cache-Control", "private, max-age=600")
			c.String(http.StatusOK, strings.ReplaceAll(CustomSwaggerInitializerJS, "{{.URL}}", swaggerDocFile))
		default:

			c.Header("Cache-Control", "private, max-age=600")
			fileServer.ServeHTTP(c.Writer, c.Request)
		}
	}
}
