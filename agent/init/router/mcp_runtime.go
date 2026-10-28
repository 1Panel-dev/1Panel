package router

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"

	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/dto/response"
	"github.com/1Panel-dev/1Panel/agent/app/service"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/middleware"
	"github.com/1Panel-dev/1Panel/agent/utils/mcpapi"
	"github.com/1Panel-dev/1Panel/agent/utils/mcpruntime"
	"github.com/gin-gonic/gin"
)

func registerMCPRuntime(engine *gin.Engine) {
	controller := mcpruntime.New()
	engine.POST("/internal/mcp/docker", mcpDockerHandler(controller))
	engine.POST("/internal/mcp/website", mcpWebsiteHandler(service.NewMCPWebsiteBindingService()))
	engine.GET(mcpapi.CapabilitiesPath, middleware.Certificate(), mcpCapabilitiesHandler())
}

type mcpWebsiteController interface {
	Handle(context.Context, request.MCPWebsiteBinding) response.MCPWebsiteBinding
}

func mcpWebsiteHandler(controller mcpWebsiteController) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		local, ok := ctx.Request.Context().Value(http.LocalAddrContextKey).(net.Addr)
		if !global.IsMaster || !ok || local.Network() != "unix" {
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}
		data, err := io.ReadAll(io.LimitReader(ctx.Request.Body, 16385))
		if err != nil || len(data) > 16384 {
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}
		var req request.MCPWebsiteBinding
		decoder := json.NewDecoder(bytes.NewReader(data))
		decoder.DisallowUnknownFields()
		if decoder.Decode(&req) != nil || decoder.Decode(new(any)) != io.EOF {
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}
		ctx.Header("Cache-Control", "no-store")
		ctx.JSON(http.StatusOK, controller.Handle(ctx.Request.Context(), req))
	}
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

type mcpDockerController interface {
	Handle(context.Context, mcpruntime.Request) (mcpruntime.Response, error)
	HandleWithProgress(context.Context, mcpruntime.Request, func(mcpruntime.PullProgress) error) (mcpruntime.Response, error)
}

func mcpDockerHandler(controller mcpDockerController) gin.HandlerFunc {
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
		if req.StreamProgress {
			if req.Operation != "pull" {
				ctx.JSON(http.StatusBadRequest, mcpruntime.Response{Error: "MCP Docker progress is only available when pulling an image"})
				return
			}
			ctx.Header("Content-Type", "application/x-ndjson")
			ctx.Header("Cache-Control", "no-store")
			ctx.Status(http.StatusOK)
			encoder := json.NewEncoder(ctx.Writer)
			_, err := controller.HandleWithProgress(ctx.Request.Context(), req, func(progress mcpruntime.PullProgress) error {
				if err := encoder.Encode(mcpruntime.PullEvent{Type: "progress", PullProgress: progress}); err != nil {
					return err
				}
				ctx.Writer.Flush()
				return nil
			})
			result := mcpruntime.PullEvent{Type: "result"}
			if err != nil {
				result.Error = err.Error()
			}
			_ = encoder.Encode(result)
			ctx.Writer.Flush()
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
