package middleware

import (
	"errors"
	"fmt"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/gin-gonic/gin"
)

func Certificate() gin.HandlerFunc {
	return func(c *gin.Context) {
		if global.IsMaster {
			c.Next()
			return
		}
		if !c.Request.TLS.HandshakeComplete || len(c.Request.TLS.PeerCertificates) == 0 {
			helper.InternalServer(c, errors.New("no such tls peer certificates"))
			return
		}
		cert := c.Request.TLS.PeerCertificates[0]
		if cert.Subject.CommonName != "panel_client" {
			helper.InternalServer(c, fmt.Errorf("err certificate"))
			return
		}
		conn := c.Request.Header.Get("Connection")
		if conn == "Upgrade" {
			c.Next()
			return
		}
		masterProxyID := c.Request.Header.Get("Proxy-Id")
		proxyID, err := cmd.RunDefaultWithStdoutBashC("cat /etc/1panel/.nodeProxyID")
		if err == nil && len(proxyID) != 0 && strings.TrimSpace(proxyID) != strings.TrimSpace(masterProxyID) {
			helper.InternalServer(c, fmt.Errorf("err proxy id"))
			return
		}
		c.Next()
	}
}
