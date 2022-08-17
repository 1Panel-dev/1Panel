package middleware

import (
	"net/http"

	"github.com/1Panel-dev/1Panel/global"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/csrf"
	adapter "github.com/gwatts/gin-adapter"
)

func CSRF() gin.HandlerFunc {
	csrfMd := csrf.Protect(
		[]byte(global.CONF.Csrf.Key),
		csrf.ErrorHandler(http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte("csrf token invalid"))
			})),
	)
	return adapter.Wrap(csrfMd)
}

func LoadCsrfToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-CSRF-TOKEN", csrf.Token(c.Request))
	}
}
