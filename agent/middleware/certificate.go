package middleware

import (
	"errors"
	"fmt"

	"github.com/1Panel-dev/1Panel/agent/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/gin-gonic/gin"
)

func Certificate() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !c.Request.TLS.HandshakeComplete || len(c.Request.TLS.PeerCertificates) == 0 {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, errors.New("no such tls peer certificates"))
			return
		}
		cert := c.Request.TLS.PeerCertificates[0]
		if cert.Subject.CommonName != "panel_client" {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, fmt.Errorf("err certificate"))
			return
		}
		c.Next()
	}
}
