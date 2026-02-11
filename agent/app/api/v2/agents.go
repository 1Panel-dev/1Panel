package v2

import (
	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/gin-gonic/gin"
)

// @Tags AI
// @Summary Create Agent
// @Accept json
// @Param request body dto.AgentCreateReq true "request"
// @Success 200 {object} dto.AgentItem
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/agents [post]
func (b *BaseApi) CreateAgent(c *gin.Context) {
	var req dto.AgentCreateReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	res, err := agentService.Create(req)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags AI
// @Summary Page Agents
// @Accept json
// @Param request body dto.SearchWithPage true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/agents/search [post]
func (b *BaseApi) PageAgents(c *gin.Context) {
	var req dto.SearchWithPage
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	total, list, err := agentService.Page(req)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.SuccessWithData(c, dto.PageResult{
		Items: list,
		Total: total,
	})
}

// @Tags AI
// @Summary Delete Agent
// @Accept json
// @Param request body dto.AgentDeleteReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/agents/delete [post]
func (b *BaseApi) DeleteAgent(c *gin.Context) {
	var req dto.AgentDeleteReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := agentService.Delete(req); err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.Success(c)
}

// @Tags AI
// @Summary Reset Agent token
// @Accept json
// @Param request body dto.AgentTokenResetReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/agents/token/reset [post]
func (b *BaseApi) ResetAgentToken(c *gin.Context) {
	var req dto.AgentTokenResetReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := agentService.ResetToken(req); err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.Success(c)
}

// @Tags AI
// @Summary Update Agent model config
// @Accept json
// @Param request body dto.AgentModelConfigUpdateReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/agents/model/update [post]
func (b *BaseApi) UpdateAgentModelConfig(c *gin.Context) {
	var req dto.AgentModelConfigUpdateReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := agentService.UpdateModelConfig(req); err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.Success(c)
}

// @Tags AI
// @Summary Get Providers
// @Success 200 {object} []dto.ProviderInfo
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/agents/providers [get]
func (b *BaseApi) GetAgentProviders(c *gin.Context) {
	list, err := agentService.GetProviders()
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.SuccessWithData(c, list)
}

// @Tags AI
// @Summary Create Agent account
// @Accept json
// @Param request body dto.AgentAccountCreateReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/agents/accounts [post]
func (b *BaseApi) CreateAgentAccount(c *gin.Context) {
	var req dto.AgentAccountCreateReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := agentService.CreateAccount(req); err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.Success(c)
}

// @Tags AI
// @Summary Update Agent account
// @Accept json
// @Param request body dto.AgentAccountUpdateReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/agents/accounts/update [post]
func (b *BaseApi) UpdateAgentAccount(c *gin.Context) {
	var req dto.AgentAccountUpdateReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := agentService.UpdateAccount(req); err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.Success(c)
}

// @Tags AI
// @Summary Page Agent accounts
// @Accept json
// @Param request body dto.AgentAccountSearch true "request"
// @Success 200 {object} dto.PageResult
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/agents/accounts/search [post]
func (b *BaseApi) PageAgentAccounts(c *gin.Context) {
	var req dto.AgentAccountSearch
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	total, list, err := agentService.PageAccounts(req)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.SuccessWithData(c, dto.PageResult{
		Items: list,
		Total: total,
	})
}

// @Tags AI
// @Summary Verify Agent account
// @Accept json
// @Param request body dto.AgentAccountVerifyReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/agents/accounts/verify [post]
func (b *BaseApi) VerifyAgentAccount(c *gin.Context) {
	var req dto.AgentAccountVerifyReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := agentService.VerifyAccount(req); err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.Success(c)
}

// @Tags AI
// @Summary Delete Agent account
// @Accept json
// @Param request body dto.AgentAccountDeleteReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/agents/accounts/delete [post]
func (b *BaseApi) DeleteAgentAccount(c *gin.Context) {
	var req dto.AgentAccountDeleteReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := agentService.DeleteAccount(req); err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.Success(c)
}

// @Tags AI
// @Summary Get Agent Feishu channel config
// @Accept json
// @Param request body dto.AgentFeishuConfigReq true "request"
// @Success 200 {object} dto.AgentFeishuConfig
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/agents/channel/feishu/get [post]
func (b *BaseApi) GetAgentFeishuConfig(c *gin.Context) {
	var req dto.AgentFeishuConfigReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	data, err := agentService.GetFeishuConfig(req)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags AI
// @Summary Update Agent Feishu channel config
// @Accept json
// @Param request body dto.AgentFeishuConfigUpdateReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/agents/channel/feishu/update [post]
func (b *BaseApi) UpdateAgentFeishuConfig(c *gin.Context) {
	var req dto.AgentFeishuConfigUpdateReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := agentService.UpdateFeishuConfig(req); err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.Success(c)
}

// @Tags AI
// @Summary Approve Agent Feishu pairing code
// @Accept json
// @Param request body dto.AgentFeishuPairingApproveReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/agents/channel/feishu/approve [post]
func (b *BaseApi) ApproveAgentFeishuPairing(c *gin.Context) {
	var req dto.AgentFeishuPairingApproveReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := agentService.ApproveFeishuPairing(req); err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.Success(c)
}
