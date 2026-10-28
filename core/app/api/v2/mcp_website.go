package v2

import (
	"context"
	"time"

	"github.com/1Panel-dev/1Panel/core/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/service"
	contract "github.com/1Panel-dev/1Panel/core/utils/mcp"
	"github.com/gin-gonic/gin"
)

func (b *BaseApi) MCPWebsiteStatus(c *gin.Context) {
	if _, _, ok := mcpManagement(c, false, true); !ok {
		return
	}
	if c.Request.URL.RawQuery != "" {
		mcpManagementError(c, contract.Invalid())
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	result, err := service.NewMCPRuntimeService().WebsiteStatus(ctx)
	if err != nil {
		mcpManagementError(c, err)
		return
	}
	helper.SuccessWithData(c, result)
}

func (b *BaseApi) BindMCPWebsite(c *gin.Context) {
	if _, _, ok := mcpManagement(c, true, true); !ok {
		return
	}
	var req dto.MCPWebsiteBind
	if !mcpRead(c, &req) {
		return
	}
	ctx, cancel := context.WithTimeout(context.WithoutCancel(c.Request.Context()), 90*time.Second)
	defer cancel()
	result, err := service.NewMCPRuntimeService().BindWebsite(ctx, req)
	if err != nil {
		mcpManagementError(c, err)
		return
	}
	helper.SuccessWithData(c, result)
}

func (b *BaseApi) UpdateMCPWebsite(c *gin.Context) {
	if _, _, ok := mcpManagement(c, true, true); !ok {
		return
	}
	var req dto.MCPWebsiteUpdate
	if !mcpRead(c, &req) {
		return
	}
	ctx, cancel := context.WithTimeout(context.WithoutCancel(c.Request.Context()), 90*time.Second)
	defer cancel()
	result, err := service.NewMCPRuntimeService().UpdateWebsite(ctx, req)
	if err != nil {
		mcpManagementError(c, err)
		return
	}
	helper.SuccessWithData(c, result)
}

func (b *BaseApi) UnbindMCPWebsite(c *gin.Context) {
	if _, _, ok := mcpManagement(c, true, true); !ok {
		return
	}
	var req dto.MCPWebsiteUnbind
	if !mcpRead(c, &req) {
		return
	}
	ctx, cancel := context.WithTimeout(context.WithoutCancel(c.Request.Context()), 90*time.Second)
	defer cancel()
	result, err := service.NewMCPRuntimeService().UnbindWebsite(ctx, req)
	if err != nil {
		mcpManagementError(c, err)
		return
	}
	helper.SuccessWithData(c, result)
}
