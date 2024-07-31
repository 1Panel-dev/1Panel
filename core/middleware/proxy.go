package middleware

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	"github.com/1Panel-dev/1Panel/core/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/xpack/utlis/proxy"
	"github.com/gin-gonic/gin"
)

func Proxy() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/v2/core") {
			c.Next()
			return
		}
		currentNode := c.Request.Header.Get("CurrentNode")
		if len(currentNode) != 0 && currentNode != "127.0.0.1" {
			if err := proxy.Proxy(c, currentNode); err != nil {
				fmt.Println(err)
				helper.ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrProxy, err)
				return
			}
			c.Abort()
			return
		}
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
	}
}
