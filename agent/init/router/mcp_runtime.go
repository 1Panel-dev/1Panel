package router

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"

	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/middleware"
	"github.com/1Panel-dev/1Panel/agent/utils/mcpapi"
	"github.com/1Panel-dev/1Panel/agent/utils/mcpruntime"
	"github.com/gin-gonic/gin"
)

func registerMCPRuntime(engine *gin.Engine) {
	controller := mcpruntime.New()
	engine.POST("/internal/mcp/docker", mcpDockerHandler(controller))
	engine.GET(mcpapi.CapabilitiesPath, middleware.Certificate(), mcpCapabilitiesHandler())
}

func mcpCapabilitiesHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !mcpapi.TrustedTransport(ctx.Request) {
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"contractVersion": mcpapi.ContractVersion, "features": []string{
			"operation-bound-identity", "stable-node-target", "resource-guards",
		}})
	}
}

func mcpDockerHandler(controller *mcpruntime.Controller) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		local, ok := ctx.Request.Context().Value(http.LocalAddrContextKey).(net.Addr)
		if !global.IsMaster || !ok || local.Network() != "unix" {
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}
		data, err := io.ReadAll(io.LimitReader(ctx.Request.Body, 8193))
		if err != nil || len(data) > 8192 {
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}
		var req mcpruntime.Request
		decoder := json.NewDecoder(bytes.NewReader(data))
		decoder.DisallowUnknownFields()
		if decoder.Decode(&req) != nil || decoder.Decode(new(any)) != io.EOF {
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}
		response, err := controller.Handle(ctx.Request.Context(), req)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, mcpruntime.Response{Error: err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, response)
	}
}
