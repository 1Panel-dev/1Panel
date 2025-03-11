package router

import (
	"context"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	"github.com/1Panel-dev/1Panel/core/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/core/utils/xpack"
	"github.com/gin-gonic/gin"
)

var wsUrl = map[string]struct{}{
	"/api/v2/process/ws":         {},
	"/api/v2/files/wget/process": {},

	"/api/v2/containers/search/log": {},
}

func Proxy() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqPath := c.Request.URL.Path
		if strings.HasPrefix(reqPath, "/1panel/swagger") || !strings.HasPrefix(c.Request.URL.Path, "/api/v2") {
			c.Next()
			return
		}
		if strings.HasPrefix(reqPath, "/api/v2/core") && !strings.HasPrefix(c.Request.URL.Path, "/api/v2/core/xpack") {
			c.Next()
			return
		}
		var currentNode string
		if _, ok := wsUrl[reqPath]; ok {
			currentNode = c.Query("currentNode")
		} else {
			currentNode = c.Request.Header.Get("CurrentNode")
		}

		if !strings.HasPrefix(c.Request.URL.Path, "/api/v2/core") && (currentNode == "local" || len(currentNode) == 0 || currentNode == "127.0.0.1") {
			sockPath := "/etc/1panel/agent.sock"
			if _, err := os.Stat(sockPath); err != nil {
				helper.ErrorWithDetail(c, http.StatusBadRequest, "ErrProxy", err)
				return
			}
			dialUnix := func() (conn net.Conn, err error) {
				return net.Dial("unix", sockPath)
			}
			transport := &http.Transport{
				DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
					return dialUnix()
				},
			}
			proxy := &httputil.ReverseProxy{
				Director: func(req *http.Request) {
					req.URL.Scheme = "http"
					req.URL.Host = "unix"
				},
				Transport: transport,
			}
			proxy.ServeHTTP(c.Writer, c.Request)
			c.Abort()
			return
		}
		xpack.Proxy(c, currentNode)
		c.Abort()
	}
}
