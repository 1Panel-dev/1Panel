package router

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/1Panel-dev/1Panel/core/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/core/cmd/server/res"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/init/proxy"
	psessionUtils "github.com/1Panel-dev/1Panel/core/init/session/psession"
	"github.com/1Panel-dev/1Panel/core/middleware"
	"github.com/1Panel-dev/1Panel/core/utils/xpack"
	"github.com/gin-gonic/gin"
)

func Proxy() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqPath := c.Request.URL.Path
		if !middleware.ShouldProxyToAgent(reqPath) {
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

		if !apiReq && !isLocalAPI(reqPath) && !isPublicFileShareAPI(reqPath) && !checkSession(c) {
			data, _ := res.ErrorMsg.ReadFile("html/401.html")
			c.Data(401, "text/html; charset=utf-8", data)
			c.Abort()
			return
		}

		if reqPath == "/api/v2/hosts/terminal" && (currentNode == "local" || len(currentNode) == 0) {
			proxyLocalAgent(c)
			return
		}

		if !strings.HasPrefix(reqPath, "/api/v2/core") && (currentNode == "local" || len(currentNode) == 0) {
			proxyLocalAgent(c)
			return
		}
		xpack.MultiNodeProvider.Proxy(c, currentNode)
		c.Abort()
	}
}

func proxyLocalAgent(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil && err != http.ErrAbortHandler {
			global.LOG.Debug(err)
		}
	}()
	proxy.LocalAgentProxy.ServeHTTP(c.Writer, c.Request)
	c.Abort()
}

func checkSession(c *gin.Context) bool {
	psession, err := global.SESSION.Get(c)
	if err != nil {
		return false
	}
	c.Set(psessionUtils.GinContextSessionUserKey, psession)
	lifeTime, err := xpack.AuthProvider.LoadSessionTimeout(c, psession)
	if err != nil {
		global.LOG.Errorf("get session timeout failed, err: %v", err)
		return false
	}
	if _, err := global.SESSION.RefreshIfNeeded(c, psession, global.CONF.Conn.SSL == constant.StatusEnable, lifeTime); err != nil {
		global.LOG.Warnf("proxy refresh session failed, path=%s, err=%v", c.Request.URL.Path, err)
		return false
	}
	return true
}

func isLocalAPI(urlPath string) bool {
	return urlPath == "/api/v2/core/xpack/sync/ssl"
}

func isPublicFileShareAPI(urlPath string) bool {
	return urlPath == "/api/v2/files/share/download" || urlPath == "/api/v2/files/share/check" || urlPath == "/api/v2/files/share/info"
}
