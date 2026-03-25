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
// @Summary Get Agent overview
// @Accept json
// @Param request body dto.AgentOverviewReq true "request"
// @Success 200 {object} dto.AgentOverview
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/agents/overview [post]
func (b *BaseApi) GetAgentOverview(c *gin.Context) {
	var req dto.AgentOverviewReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	res, err := agentService.GetOverview(req)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.SuccessWithData(c, res)
}

// @Tags AI
// @Summary Get Providers
// @Success 200 {array} dto.ProviderInfo
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
// @Summary List Agent account models
// @Accept json
// @Param request body dto.AgentAccountModelReq true "request"
// @Success 200 {array} dto.AgentAccountModel
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/agents/accounts/models [post]
func (b *BaseApi) GetAgentAccountModels(c *gin.Context) {
	var req dto.AgentAccountModelReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	list, err := agentService.GetAccountModels(req)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.SuccessWithData(c, list)
}

// @Tags AI
// @Summary Create Agent account model
// @Accept json
// @Param request body dto.AgentAccountModelCreateReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/agents/accounts/models/create [post]
func (b *BaseApi) CreateAgentAccountModel(c *gin.Context) {
	var req dto.AgentAccountModelCreateReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := agentService.CreateAccountModel(req); err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.Success(c)
}

// @Tags AI
// @Summary Update Agent account model
// @Accept json
// @Param request body dto.AgentAccountModelUpdateReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/agents/accounts/models/update [post]
func (b *BaseApi) UpdateAgentAccountModel(c *gin.Context) {
	var req dto.AgentAccountModelUpdateReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := agentService.UpdateAccountModel(req); err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.Success(c)
}

// @Tags AI
// @Summary Delete Agent account model
// @Accept json
// @Param request body dto.AgentAccountModelDeleteReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/agents/accounts/models/delete [post]
func (b *BaseApi) DeleteAgentAccountModel(c *gin.Context) {
	var req dto.AgentAccountModelDeleteReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := agentService.DeleteAccountModel(req); err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.Success(c)
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
// @Summary Get Agent Telegram channel config
// @Accept json
// @Param request body dto.AgentTelegramConfigReq true "request"
// @Success 200 {object} dto.AgentTelegramConfig
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/agents/channel/telegram/get [post]
func (b *BaseApi) GetAgentTelegramConfig(c *gin.Context) {
	var req dto.AgentTelegramConfigReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	data, err := agentService.GetTelegramConfig(req)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags AI
// @Summary Update Agent Telegram channel config
// @Accept json
// @Param request body dto.AgentTelegramConfigUpdateReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/agents/channel/telegram/update [post]
func (b *BaseApi) UpdateAgentTelegramConfig(c *gin.Context) {
	var req dto.AgentTelegramConfigUpdateReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := agentService.UpdateTelegramConfig(req); err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.Success(c)
}

// @Tags AI
// @Summary Get Agent Discord channel config
// @Accept json
// @Param request body dto.AgentDiscordConfigReq true "request"
// @Success 200 {object} dto.AgentDiscordConfig
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/agents/channel/discord/get [post]
func (b *BaseApi) GetAgentDiscordConfig(c *gin.Context) {
	var req dto.AgentDiscordConfigReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	data, err := agentService.GetDiscordConfig(req)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags AI
// @Summary Update Agent Discord channel config
// @Accept json
// @Param request body dto.AgentDiscordConfigUpdateReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/agents/channel/discord/update [post]
func (b *BaseApi) UpdateAgentDiscordConfig(c *gin.Context) {
	var req dto.AgentDiscordConfigUpdateReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := agentService.UpdateDiscordConfig(req); err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.Success(c)
}

// @Tags AI
// @Summary Get Agent QQ Bot channel config
// @Accept json
// @Param request body dto.AgentWecomConfigReq true "request"
// @Success 200 {object} dto.AgentWecomConfig
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/agents/channel/wecom/get [post]
func (b *BaseApi) GetAgentWecomConfig(c *gin.Context) {
	var req dto.AgentWecomConfigReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	data, err := agentService.GetWecomConfig(req)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags AI
// @Summary Update Agent WeCom channel config
// @Accept json
// @Param request body dto.AgentWecomConfigUpdateReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/agents/channel/wecom/update [post]
func (b *BaseApi) UpdateAgentWecomConfig(c *gin.Context) {
	var req dto.AgentWecomConfigUpdateReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := agentService.UpdateWecomConfig(req); err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.Success(c)
}

// @Tags AI
// @Summary Get Agent DingTalk channel config
// @Accept json
// @Param request body dto.AgentDingTalkConfigReq true "request"
// @Success 200 {object} dto.AgentDingTalkConfig
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/agents/channel/dingtalk/get [post]
func (b *BaseApi) GetAgentDingTalkConfig(c *gin.Context) {
	var req dto.AgentDingTalkConfigReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	data, err := agentService.GetDingTalkConfig(req)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags AI
// @Summary Update Agent DingTalk channel config
// @Accept json
// @Param request body dto.AgentDingTalkConfigUpdateReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/agents/channel/dingtalk/update [post]
func (b *BaseApi) UpdateAgentDingTalkConfig(c *gin.Context) {
	var req dto.AgentDingTalkConfigUpdateReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := agentService.UpdateDingTalkConfig(req); err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.Success(c)
}

// @Tags AI
// @Summary Get Agent QQ Bot channel config
// @Accept json
// @Param request body dto.AgentQQBotConfigReq true "request"
// @Success 200 {object} dto.AgentQQBotConfig
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/agents/channel/qqbot/get [post]
func (b *BaseApi) GetAgentQQBotConfig(c *gin.Context) {
	var req dto.AgentQQBotConfigReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	data, err := agentService.GetQQBotConfig(req)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags AI
// @Summary Update Agent QQ Bot channel config
// @Accept json
// @Param request body dto.AgentQQBotConfigUpdateReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/agents/channel/qqbot/update [post]
func (b *BaseApi) UpdateAgentQQBotConfig(c *gin.Context) {
	var req dto.AgentQQBotConfigUpdateReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := agentService.UpdateQQBotConfig(req); err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.Success(c)
}

// @Tags AI
// @Summary Install Agent plugin
// @Accept json
// @Param request body dto.AgentPluginInstallReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/agents/plugin/install [post]
func (b *BaseApi) InstallAgentPlugin(c *gin.Context) {
	var req dto.AgentPluginInstallReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := agentService.InstallPlugin(req); err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.Success(c)
}

// @Tags AI
// @Summary Check Agent plugin installation status
// @Accept json
// @Param request body dto.AgentPluginCheckReq true "request"
// @Success 200 {object} dto.AgentPluginStatus
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/agents/plugin/check [post]
func (b *BaseApi) CheckAgentPlugin(c *gin.Context) {
	var req dto.AgentPluginCheckReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	data, err := agentService.CheckPlugin(req)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags AI
// @Summary Get Agent Security config
// @Accept json
// @Param request body dto.AgentSecurityConfigReq true "request"
// @Success 200 {object} dto.AgentSecurityConfig
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/agents/security/get [post]
func (b *BaseApi) GetAgentSecurityConfig(c *gin.Context) {
	var req dto.AgentSecurityConfigReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	data, err := agentService.GetSecurityConfig(req)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags AI
// @Summary Update Agent Security config
// @Accept json
// @Param request body dto.AgentSecurityConfigUpdateReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/agents/security/update [post]
func (b *BaseApi) UpdateAgentSecurityConfig(c *gin.Context) {
	var req dto.AgentSecurityConfigUpdateReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := agentService.UpdateSecurityConfig(req); err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.Success(c)
}

// @Tags AI
// @Summary Get Agent Other config
// @Accept json
// @Param request body dto.AgentOtherConfigReq true "request"
// @Success 200 {object} dto.AgentOtherConfig
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/agents/other/get [post]
func (b *BaseApi) GetAgentOtherConfig(c *gin.Context) {
	var req dto.AgentOtherConfigReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	data, err := agentService.GetOtherConfig(req)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags AI
// @Summary Update Agent Other config
// @Accept json
// @Param request body dto.AgentOtherConfigUpdateReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/agents/other/update [post]
func (b *BaseApi) UpdateAgentOtherConfig(c *gin.Context) {
	var req dto.AgentOtherConfigUpdateReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := agentService.UpdateOtherConfig(req); err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.Success(c)
}

// @Tags AI
// @Summary Get Agent config file
// @Accept json
// @Param request body dto.AgentConfigFileReq true "request"
// @Success 200 {object} dto.AgentConfigFile
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/agents/config-file/get [post]
func (b *BaseApi) GetAgentConfigFile(c *gin.Context) {
	var req dto.AgentConfigFileReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	data, err := agentService.GetConfigFile(req)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags AI
// @Summary Update Agent config file
// @Accept json
// @Param request body dto.AgentConfigFileUpdateReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/agents/config-file/update [post]
func (b *BaseApi) UpdateAgentConfigFile(c *gin.Context) {
	var req dto.AgentConfigFileUpdateReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := agentService.UpdateConfigFile(req); err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.Success(c)
}

// @Tags AI
// @Summary List Agent skills
// @Accept json
// @Param request body dto.AgentSkillsReq true "request"
// @Success 200 {array} dto.AgentSkillItem
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/agents/skills/list [post]
func (b *BaseApi) ListAgentSkills(c *gin.Context) {
	var req dto.AgentSkillsReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	data, err := agentService.ListSkills(req)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags AI
// @Summary Search Agent skills
// @Accept json
// @Param request body dto.AgentSkillSearchReq true "request"
// @Success 200 {array} dto.AgentSkillSearchItem
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/agents/skills/search [post]
func (b *BaseApi) SearchAgentSkills(c *gin.Context) {
	var req dto.AgentSkillSearchReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	data, err := agentService.SearchSkills(req)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags AI
// @Summary Update Agent skill status
// @Accept json
// @Param request body dto.AgentSkillUpdateReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/agents/skills/update [post]
func (b *BaseApi) UpdateAgentSkill(c *gin.Context) {
	var req dto.AgentSkillUpdateReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := agentService.UpdateSkill(req); err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.Success(c)
}

// @Tags AI
// @Summary Install Agent skill
// @Accept json
// @Param request body dto.AgentSkillInstallReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/agents/skills/install [post]
func (b *BaseApi) InstallAgentSkill(c *gin.Context) {
	var req dto.AgentSkillInstallReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := agentService.InstallSkill(req); err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.Success(c)
}

// @Tags AI
// @Summary Login Agent Weixin channel
// @Accept json
// @Param request body dto.AgentWeixinLoginReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/agents/channel/weixin/login [post]
func (b *BaseApi) LoginAgentWeixinChannel(c *gin.Context) {
	var req dto.AgentWeixinLoginReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := agentService.LoginWeixinChannel(req); err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.Success(c)
}

// @Tags AI
// @Summary Approve Agent channel pairing code
// @Accept json
// @Param request body dto.AgentChannelPairingApproveReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /ai/agents/channel/pairing/approve [post]
func (b *BaseApi) ApproveAgentChannelPairing(c *gin.Context) {
	var req dto.AgentChannelPairingApproveReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := agentService.ApproveChannelPairing(req); err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.Success(c)
}
