package router

import (
	"context"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/cmd/server/res"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/utils/security"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
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

		apiReq := c.GetBool("API_AUTH")

		if !apiReq && strings.HasPrefix(c.Request.URL.Path, "/api/v2/") && !checkSession(c) {
			data, _ := res.ErrorMsg.ReadFile("html/401.html")
			c.Data(401, "text/html; charset=utf-8", data)
			c.Abort()
			return
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
				ModifyResponse: func(response *http.Response) error {
					if response.StatusCode == 404 {
						security.HandleNotSecurity(c, "")
						c.Abort()
					}
					return nil
				},
			}
			proxy.ServeHTTP(c.Writer, c.Request)
			c.Abort()
			return
		}
		xpack.Proxy(c, currentNode)
		c.Abort()
	}
}

func checkSession(c *gin.Context) bool {
	psession, err := global.SESSION.Get(c)
	if err != nil {
		return false
	}
	settingRepo := repo.NewISettingRepo()
	setting, err := settingRepo.Get(repo.WithByKey("SessionTimeout"))
	if err != nil {
		return false
	}
	lifeTime, _ := strconv.Atoi(setting.Value)
	httpsSetting, err := settingRepo.Get(repo.WithByKey("SSL"))
	if err != nil {
		return false
	}
	_ = global.SESSION.Set(c, psession, httpsSetting.Value == constant.StatusEnable, lifeTime)
	return true
}
