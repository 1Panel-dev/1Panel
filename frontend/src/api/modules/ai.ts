import { AI } from '@/api/interface/ai';
import http from '@/api';
import { ResPage, SearchWithPage } from '../interface';
import { TimeoutEnum } from '@/enums/http-enum';

export const createOllamaModel = (name: string, taskID: string) => {
    return http.post(`/ai/ollama/model`, { name: name, taskID: taskID });
};
export const recreateOllamaModel = (name: string, taskID: string) => {
    return http.post(`/ai/ollama/model/recreate`, { name: name, taskID: taskID });
};
export const deleteOllamaModel = (ids: Array<number>, force: boolean) => {
    return http.post(`/ai/ollama/model/del`, { ids: ids, forceDelete: force });
};
export const searchOllamaModel = (params: AI.OllamaModelSearch) => {
    return http.post<ResPage<AI.OllamaModelInfo>>(`/ai/ollama/model/search`, params);
};
export const loadOllamaModel = (name: string) => {
    return http.post<string>(`/ai/ollama/model/load`, { name: name });
};
export const syncOllamaModel = () => {
    return http.post<Array<AI.OllamaModelDropInfo>>(`/ai/ollama/model/sync`);
};
export const closeOllamaModel = (name: string) => {
    return http.post(`/ai/ollama/close`, { name: name });
};

export const loadGPUInfo = () => {
    return http.get<any>(`/ai/gpu/load`);
};

export const bindDomain = (req: AI.BindDomain) => {
    return http.post(`/ai/domain/bind`, req);
};

export const getBindDomain = (req: AI.BindDomainReq) => {
    return http.post<AI.BindDomainRes>(`/ai/domain/get`, req);
};

export const updateBindDomain = (req: AI.BindDomain) => {
    return http.post(`/ai/domain/update`, req);
};

export const pageMcpServer = (req: AI.McpServerSearch) => {
    return http.post<ResPage<AI.McpServer>>(`/ai/mcp/search`, req);
};

export const createMcpServer = (req: AI.McpServer) => {
    return http.post(`/ai/mcp/server`, req);
};

export const updateMcpServer = (req: AI.McpServer) => {
    return http.post(`/ai/mcp/server/update`, req);
};

export const deleteMcpServer = (req: AI.McpServerDelete) => {
    return http.post(`/ai/mcp/server/del`, req);
};

export const operateMcpServer = (req: AI.McpServerOperate) => {
    return http.post(`/ai/mcp/server/op`, req);
};

export const bindMcpDomain = (req: AI.McpBindDomain) => {
    return http.post(`/ai/mcp/domain/bind`, req);
};

export const getMcpDomain = () => {
    return http.get<AI.McpDomainRes>(`/ai/mcp/domain/get`);
};

export const updateMcpDomain = (req: AI.McpBindDomainUpdate) => {
    return http.post(`/ai/mcp/domain/update`, req);
};

export const pageTensorRTLLM = (req: AI.TensorRTLLMSearch) => {
    return http.post<ResPage<AI.TensorRTLLMDTO>>(`/ai/tensorrt/search`, req);
};

export const createTensorRTLLM = (req: AI.TensorRTLLM) => {
    return http.post(`/ai/tensorrt/create`, req);
};

export const updateTensorRTLLM = (req: AI.TensorRTLLM) => {
    return http.post(`/ai/tensorrt/update`, req);
};

export const deleteTensorRTLLM = (req: AI.TensorRTLLMDelete) => {
    return http.post(`/ai/tensorrt/delete`, req);
};

export const operateTensorRTLLM = (req: AI.TensorRTLLMOperate) => {
    return http.post(`/ai/tensorrt/operate`, req);
};

export const createAgent = (req: AI.AgentCreateReq) => {
    return http.post<AI.AgentItem>(`/ai/agents`, req, TimeoutEnum.T_60S);
};

export const pageAgents = (req: SearchWithPage) => {
    return http.post<ResPage<AI.AgentItem>>(`/ai/agents/search`, req);
};

export const deleteAgent = (req: AI.AgentDeleteReq) => {
    return http.post(`/ai/agents/delete`, req);
};

export const resetAgentToken = (req: AI.AgentTokenResetReq) => {
    return http.post(`/ai/agents/token/reset`, req);
};

export const updateAgentRemark = (req: AI.AgentRemarkUpdateReq) => {
    return http.post(`/ai/agents/remark`, req);
};

export const getAgentModelConfig = (req: AI.AgentIDReq) => {
    return http.post<AI.AgentModelConfig>(`/ai/agents/model/get`, req);
};

export const updateAgentModelConfig = (req: AI.AgentModelConfigUpdateReq) => {
    return http.post(`/ai/agents/model/update`, req);
};

export const getAgentOverview = (req: AI.AgentOverviewReq) => {
    return http.post<AI.AgentOverview>(`/ai/agents/overview`, req);
};

export const createAgentRole = (req: AI.AgentRoleCreateReq) => {
    return http.post<AI.AgentRoleCreateResp>(`/ai/agents/agent/create`, req);
};

export const deleteAgentRole = (req: AI.AgentRoleDeleteReq) => {
    return http.post(`/ai/agents/agent/delete`, req);
};

export const getConfiguredAgentRoles = (req: AI.AgentConfiguredAgentsReq) => {
    return http.post<AI.AgentConfiguredAgentItem[]>(`/ai/agents/agent/list`, req);
};

export const getAgentRoleChannels = (req: AI.AgentRoleChannelsReq) => {
    return http.post<AI.AgentRoleChannelItem[]>(`/ai/agents/agent/channels`, req);
};

export const getAgentRoleMarkdownFiles = (req: AI.AgentRoleMarkdownFilesReq) => {
    return http.post<AI.AgentRoleMarkdownFileItem[]>(`/ai/agents/agent/md/list`, req);
};

export const updateAgentRoleMarkdownFile = (req: AI.AgentRoleMarkdownFilesUpdateReq) => {
    return http.post(`/ai/agents/agent/md/update`, req);
};

export const getAgentProviders = () => {
    return http.get<AI.ProviderInfo[]>(`/ai/agents/providers`);
};

export const createAgentAccount = (req: AI.AgentAccountCreateReq) => {
    return http.post(`/ai/agents/accounts`, req);
};

export const updateAgentAccount = (req: AI.AgentAccountUpdateReq) => {
    return http.post(`/ai/agents/accounts/update`, req);
};

export const pageAgentAccounts = (req: AI.AgentAccountSearch) => {
    return http.post<ResPage<AI.AgentAccountItem>>(`/ai/agents/accounts/search`, req);
};

export const getAgentAccountModels = (req: AI.AgentAccountModelReq) => {
    return http.post<AI.AgentAccountModel[]>(`/ai/agents/accounts/models`, req);
};

export const createAgentAccountModel = (req: AI.AgentAccountModelCreateReq) => {
    return http.post(`/ai/agents/accounts/models/create`, req);
};

export const updateAgentAccountModel = (req: AI.AgentAccountModelUpdateReq) => {
    return http.post(`/ai/agents/accounts/models/update`, req);
};

export const deleteAgentAccountModel = (req: AI.AgentAccountModelDeleteReq) => {
    return http.post(`/ai/agents/accounts/models/delete`, req);
};

export const verifyAgentAccount = (req: AI.AgentAccountVerifyReq) => {
    return http.post(`/ai/agents/accounts/verify`, req);
};

export const deleteAgentAccount = (req: AI.AgentAccountDeleteReq) => {
    return http.post(`/ai/agents/accounts/delete`, req);
};

export const getAgentFeishuConfig = (req: AI.AgentFeishuConfigReq) => {
    return http.post<AI.AgentFeishuConfig>(`/ai/agents/channel/feishu/get`, req);
};

export const updateAgentFeishuConfig = (req: AI.AgentFeishuConfigUpdateReq) => {
    return http.post(`/ai/agents/channel/feishu/update`, req);
};

export const getAgentTelegramConfig = (req: AI.AgentTelegramConfigReq) => {
    return http.post<AI.AgentTelegramConfig>(`/ai/agents/channel/telegram/get`, req);
};

export const updateAgentTelegramConfig = (req: AI.AgentTelegramConfigUpdateReq) => {
    return http.post(`/ai/agents/channel/telegram/update`, req);
};

export const getAgentDiscordConfig = (req: AI.AgentDiscordConfigReq) => {
    return http.post<AI.AgentDiscordConfig>(`/ai/agents/channel/discord/get`, req);
};

export const updateAgentDiscordConfig = (req: AI.AgentDiscordConfigUpdateReq) => {
    return http.post(`/ai/agents/channel/discord/update`, req);
};

export const getAgentWecomConfig = (req: AI.AgentWecomConfigReq) => {
    return http.post<AI.AgentWecomConfig>(`/ai/agents/channel/wecom/get`, req);
};

export const updateAgentWecomConfig = (req: AI.AgentWecomConfigUpdateReq) => {
    return http.post(`/ai/agents/channel/wecom/update`, req);
};

export const getAgentDingTalkConfig = (req: AI.AgentDingTalkConfigReq) => {
    return http.post<AI.AgentDingTalkConfig>(`/ai/agents/channel/dingtalk/get`, req);
};

export const updateAgentDingTalkConfig = (req: AI.AgentDingTalkConfigUpdateReq) => {
    return http.post(`/ai/agents/channel/dingtalk/update`, req);
};

export const loginAgentWeixinChannel = (req: AI.AgentWeixinLoginReq) => {
    return http.post(`/ai/agents/channel/weixin/login`, req);
};

export const getAgentQQBotConfig = (req: AI.AgentQQBotConfigReq) => {
    return http.post<AI.AgentQQBotConfig>(`/ai/agents/channel/qqbot/get`, req);
};

export const updateAgentQQBotConfig = (req: AI.AgentQQBotConfigUpdateReq) => {
    return http.post(`/ai/agents/channel/qqbot/update`, req);
};

export const installAgentPlugin = (req: AI.AgentPluginInstallReq) => {
    return http.post(`/ai/agents/plugin/install`, req);
};

export const upgradeAgentPlugin = (req: AI.AgentPluginUpgradeReq) => {
    return http.post(`/ai/agents/plugin/upgrade`, req);
};

export const uninstallAgentPlugin = (req: AI.AgentPluginUninstallReq) => {
    return http.post(`/ai/agents/plugin/uninstall`, req);
};

export const checkAgentPlugin = (req: AI.AgentPluginCheckReq) => {
    return http.post<AI.AgentPluginStatus>(`/ai/agents/plugin/check`, req);
};

export const getAgentSecurityConfig = (req: AI.AgentSecurityConfigReq) => {
    return http.post<AI.AgentSecurityConfig>(`/ai/agents/security/get`, req);
};

export const updateAgentSecurityConfig = (req: AI.AgentSecurityConfigUpdateReq) => {
    return http.post(`/ai/agents/security/update`, req);
};

export const getAgentOtherConfig = (req: AI.AgentOtherConfigReq) => {
    return http.post<AI.AgentOtherConfig>(`/ai/agents/other/get`, req);
};

export const updateAgentOtherConfig = (req: AI.AgentOtherConfigUpdateReq) => {
    return http.post(`/ai/agents/other/update`, req);
};

export const getAgentConfigFile = (req: AI.AgentConfigFileReq) => {
    return http.post<AI.AgentConfigFile>(`/ai/agents/config-file/get`, req);
};

export const updateAgentConfigFile = (req: AI.AgentConfigFileUpdateReq) => {
    return http.post(`/ai/agents/config-file/update`, req);
};

export const listAgentSkills = (req: AI.AgentSkillsReq) => {
    return http.post<AI.AgentSkillItem[]>(`/ai/agents/skills/list`, req);
};

export const searchAgentSkills = (req: AI.AgentSkillSearchReq) => {
    return http.post<AI.AgentSkillSearchItem[]>(`/ai/agents/skills/search`, req);
};

export const updateAgentSkill = (req: AI.AgentSkillUpdateReq) => {
    return http.post(`/ai/agents/skills/update`, req);
};

export const installAgentSkill = (req: AI.AgentSkillInstallReq) => {
    return http.post(`/ai/agents/skills/install`, req);
};

export const approveAgentChannelPairing = (req: AI.AgentChannelPairingApproveReq) => {
    return http.post(`/ai/agents/channel/pairing/approve`, req);
};
