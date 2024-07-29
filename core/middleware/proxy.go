package middleware

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/1Panel-dev/1Panel/core/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/gin-gonic/gin"
)

func Proxy() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/v2/core") {
			c.Next()
			return
		}
		currentNode := c.Request.Header.Get("CurrentNode")
		if len(currentNode) == 0 || currentNode == "127.0.0.1" {
			sockPath := "/tmp/agent.sock"
			if _, err := os.Stat(sockPath); err != nil {
				helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrProxy, err)
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
		target, err := url.Parse(fmt.Sprintf("http://%s:9999", currentNode))
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrProxy, err)
			return
		}
		proxy := httputil.NewSingleHostReverseProxy(target)
		c.Request.Host = target.Host
		c.Request.URL.Scheme = target.Scheme
		c.Request.URL.Host = target.Host
		proxy.ServeHTTP(c.Writer, c.Request)
		c.Abort()
	}
}
