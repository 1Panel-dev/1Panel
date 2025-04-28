package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/gin-gonic/gin"
)

// @Tags McpServer
// @Summary List mcp servers
// @Accept json
// @Param request body request.McpServerSearch true "request"
// @Success 200 {object} response.McpServersRes
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /mcp/search [post]
func (b *BaseApi) PageMcpServers(c *gin.Context) {
	var req request.McpServerSearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	list := mcpServerService.Page(req)
	helper.SuccessWithData(c, list)
}

// @Tags McpServer
// @Summary Create mcp server
// @Accept json
// @Param request body request.McpServerCreate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /mcp/server [post]
func (b *BaseApi) CreateMcpServer(c *gin.Context) {
	var req request.McpServerCreate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := mcpServerService.Create(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags McpServer
// @Summary Update mcp server
// @Accept json
// @Param request body request.McpServerUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /mcp/server/update [post]
func (b *BaseApi) UpdateMcpServer(c *gin.Context) {
	var req request.McpServerUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := mcpServerService.Update(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags McpServer
// @Summary Delete mcp server
// @Accept json
// @Param request body request.McpServerDelete true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /mcp/server/del [post]
func (b *BaseApi) DeleteMcpServer(c *gin.Context) {
	var req request.McpServerDelete
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := mcpServerService.Delete(req.ID)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags McpServer
// @Summary Operate mcp server
// @Accept json
// @Param request body request.McpServerOperate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /mcp/server/op [post]
func (b *BaseApi) OperateMcpServer(c *gin.Context) {
	var req request.McpServerOperate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := mcpServerService.Operate(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags McpServer
// @Summary Bind Domain for mcp server
// @Accept json
// @Param request body request.McpBindDomain true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /mcp/domain/bind [post]
func (b *BaseApi) BindMcpDomain(c *gin.Context) {
	var req request.McpBindDomain
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := mcpServerService.BindDomain(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags McpServer
// @Summary Update bind Domain for mcp server
// @Accept json
// @Param request body request.McpBindDomainUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /mcp/domain/update [post]
func (b *BaseApi) UpdateMcpBindDomain(c *gin.Context) {
	var req request.McpBindDomainUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	err := mcpServerService.UpdateBindDomain(req)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags McpServer
// @Summary Get bin Domain for mcp server
// @Accept json
// @Success 200 {object} response.McpBindDomainRes
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /mcp/domain/get [get]
func (b *BaseApi) GetMcpBindDomain(c *gin.Context) {
	res, err := mcpServerService.GetBindDomain()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, res)
}
