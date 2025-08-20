package router

import (
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/core/init/proxy"

	"github.com/1Panel-dev/1Panel/core/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/cmd/server/res"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/utils/xpack"
	"github.com/gin-gonic/gin"
)

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
		var nodeItem string
		queryNode := c.Query("operateNode")
		if queryNode != "" && queryNode != "undefined" {
			nodeItem = queryNode
		} else {
			nodeItem = c.Request.Header.Get("CurrentNode")
		}
		currentNode, err := url.QueryUnescape(nodeItem)
		if err != nil {
			helper.ErrorWithDetail(c, http.StatusBadRequest, "ErrProxy", err)
			return
		}

		apiReq := c.GetBool("API_AUTH")

		if !apiReq && strings.HasPrefix(c.Request.URL.Path, "/api/v2/") && !isLocalAPI(c.Request.URL.Path) && !checkSession(c) {
			data, _ := res.ErrorMsg.ReadFile("html/401.html")
			c.Data(401, "text/html; charset=utf-8", data)
			c.Abort()
			return
		}

		if !strings.HasPrefix(c.Request.URL.Path, "/api/v2/core") && (currentNode == "local" || len(currentNode) == 0) {
			sockPath := "/etc/1panel/agent.sock"
			if _, err := os.Stat(sockPath); err != nil {
				helper.ErrorWithDetail(c, http.StatusBadRequest, "ErrProxy", err)
				return
			}
			defer func() {
				if err := recover(); err != nil && err != http.ErrAbortHandler {
					global.LOG.Debug(err)
				}
			}()
			proxy.LocalAgentProxy.ServeHTTP(c.Writer, c.Request)
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

func isLocalAPI(urlPath string) bool {
	return urlPath == "/api/v2/core/xpack/sync/ssl"
}
