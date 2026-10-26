package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/mcpapi"
	"github.com/1Panel-dev/1Panel/agent/utils/xpack"
	"github.com/gin-gonic/gin"
)

var loadNodeProxyID = func() (string, error) {
	return cmd.NewCommandMgr(cmd.WithTimeout(20*time.Second)).RunWithStdout("cat", "/etc/1panel/.nodeProxyID")
}

func Certificate() gin.HandlerFunc {
	return func(c *gin.Context) {
		if global.IsMaster {
			c.Next()
			return
		}
		if !xpack.MultiNodeProvider.ValidateCertificate(c) {
			CloseDirectly(c)
			return
		}
		conn := c.Request.Header.Get("Connection")
		mcpRequest, _ := mcpapi.Check(c.Request)
		mcpTransport := mcpRequest || c.Request.URL.Path == mcpapi.CapabilitiesPath
		if conn == "Upgrade" && !mcpTransport {
			c.Next()
			return
		}
		masterProxyID := c.Request.Header.Get("Proxy-Id")
		proxyID, err := loadNodeProxyID()
		if mcpTransport && (err != nil || strings.TrimSpace(proxyID) == "" || strings.TrimSpace(proxyID) != strings.TrimSpace(masterProxyID)) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		if err == nil && len(proxyID) != 0 && strings.TrimSpace(proxyID) != strings.TrimSpace(masterProxyID) {
			helper.InternalServer(c, fmt.Errorf("err proxy id"))
			return
		}
		c.Request = mcpapi.BindVerifiedNodeTransport(c.Request)
		c.Next()
	}
}

func CloseDirectly(c *gin.Context) {
	hijacker, ok := c.Writer.(http.Hijacker)
	if !ok {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	conn, _, err := hijacker.Hijack()
	if err != nil {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	_ = conn.(*net.TCPConn).SetLinger(0)
	conn.Close()
}
