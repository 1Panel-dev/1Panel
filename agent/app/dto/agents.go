package dto

import "time"

type AgentCreateReq struct {
	Name           string   `json:"name" validate:"required"`
	Remark         string   `json:"remark"`
	AppVersion     string   `json:"appVersion" validate:"required"`
	WebUIPort      int      `json:"webUIPort" validate:"required,min=1,max=65535"`
	BridgePort     int      `json:"bridgePort"`
	AllowedOrigins []string `json:"allowedOrigins"`
	AgentType      string   `json:"agentType" validate:"required,oneof=openclaw copaw hermes-agent"`
	Model          string   `json:"model"`
	AccountID      uint     `json:"accountId"`
	Token          string   `json:"token"`
	TaskID         string   `json:"taskID"`
	Advanced       bool     `json:"advanced"`
	ContainerName  string   `json:"containerName"`
	AllowPort      bool     `json:"allowPort"`
	SpecifyIP      string   `json:"specifyIP"`
	RestartPolicy  string   `json:"restartPolicy"`
	CpuQuota       float64  `json:"cpuQuota"`
	MemoryLimit    float64  `json:"memoryLimit"`
	MemoryUnit     string   `json:"memoryUnit"`
	PullImage      bool     `json:"pullImage"`
	EditCompose    bool     `json:"editCompose"`
	DockerCompose  string   `json:"dockerCompose"`
}

type AgentItem struct {
	ID                   uint      `json:"id"`
	Name                 string    `json:"name"`
	Remark               string    `json:"remark"`
	AgentType            string    `json:"agentType"`
	Provider             string    `json:"provider"`
	ProviderName         string    `json:"providerName"`
	Model                string    `json:"model"`
	APIType              string    `json:"apiType"`
	MaxTokens            int       `json:"maxTokens"`
	ContextWindow        int       `json:"contextWindow"`
	BaseURL              string    `json:"baseUrl"`
	APIKey               string    `json:"apiKey"`
	Token                string    `json:"token"`
	Status               string    `json:"status"`
	Message              string    `json:"message"`
	AppInstallID         uint      `json:"appInstallId"`
	WebsiteID            uint      `json:"websiteId"`
	WebsitePrimaryDomain string    `json:"websitePrimaryDomain"`
	WebsiteType          string    `json:"websiteType"`
	WebsiteProtocol      string    `json:"websiteProtocol"`
	AccountID            uint      `json:"accountId"`
	AppVersion           string    `json:"appVersion"`
	Container            string    `json:"containerName"`
	WebUIPort            int       `json:"webUIPort"`
	BridgePort           int       `json:"bridgePort"`
	Path                 string    `json:"path"`
	ConfigPath           string    `json:"configPath"`
	Upgradable           bool      `json:"upgradable"`
	CreatedAt            time.Time `json:"createdAt"`
}

type AgentDeleteReq struct {
	ID          uint   `json:"id" validate:"required"`
	TaskID      string `json:"taskID"`
	ForceDelete bool   `json:"forceDelete"`
}

type AgentTokenResetReq struct {
	ID uint `json:"id" validate:"required"`
}

type AgentRemarkUpdateReq struct {
	ID     uint   `json:"id" validate:"required"`
	Remark string `json:"remark"`
}

type AgentWebsiteBindReq struct {
	AgentID   uint `json:"agentId" validate:"required"`
	WebsiteID uint `json:"websiteId" validate:"required"`
}

type AgentModelConfigUpdateReq struct {
	AgentID   uint     `json:"agentId" validate:"required"`
	AccountID uint     `json:"accountId" validate:"required"`
	Model     string   `json:"model" validate:"required"`
	Fallbacks []string `json:"fallbacks"`
}

type AgentModelConfig struct {
	AccountID uint     `json:"accountId"`
	Model     string   `json:"model"`
	Fallbacks []string `json:"fallbacks"`
}

type AgentHermesChatSessionItem struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Model        string `json:"model"`
	MessageCount int64  `json:"messageCount"`
	StartedAt    string `json:"startedAt"`
	LastActive   string `json:"lastActive"`
}

type AgentOverviewReq struct {
	AgentID uint `json:"agentId" validate:"required"`
}

type AgentIDReq struct {
	AgentID uint `json:"agentId" validate:"required"`
}

type AgentHermesChatSessionRenameReq struct {
	AgentID uint   `json:"agentId" validate:"required"`
	ID      string `json:"id" validate:"required"`
	Title   string `json:"title" validate:"required"`
}

type AgentOverview struct {
	Snapshot AgentOverviewSnapshot `json:"snapshot"`
}

type AgentRoleBinding struct {
	Channel   string `json:"channel" validate:"required"`
	AccountID string `json:"accountId"`
}

type AgentRoleCreateReq struct {
	AgentID  uint               `json:"agentId" validate:"required"`
	Name     string             `json:"name" validate:"required"`
	Model    string             `json:"model"`
	Bindings []AgentRoleBinding `json:"bindings"`
}

type AgentRoleCreateResp struct {
	Output string `json:"output"`
}

type AgentRoleDeleteReq struct {
	AgentID uint   `json:"agentId" validate:"required"`
	ID      string `json:"id" validate:"required"`
}

type AgentRoleBindReq struct {
	AgentID   uint   `json:"agentId" validate:"required"`
	ID        string `json:"id" validate:"required"`
	Channel   string `json:"channel" validate:"required"`
	AccountID string `json:"accountId"`
}

type AgentConfiguredAgentsReq struct {
	AgentID uint `json:"agentId" validate:"required"`
}

type AgentRoleChannelsReq struct {
	AgentID uint `json:"agentId" validate:"required"`
}

type AgentRoleChannelItem struct {
	Name       string   `json:"name"`
	Bound      bool     `json:"bound"`
	AccountIDs []string `json:"accountIds"`
}

type AgentRoleMarkdownFilesReq struct {
	AgentID   uint   `json:"agentId" validate:"required"`
	Workspace string `json:"workspace" validate:"required"`
}

type AgentConfiguredAgentItem struct {
	ID        string             `json:"id"`
	Name      string             `json:"name"`
	Workspace string             `json:"workspace"`
	Model     string             `json:"model"`
	AgentDir  string             `json:"agentDir"`
	Bindings  []AgentRoleBinding `json:"bindings"`
}

type AgentRoleMarkdownFileItem struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type AgentRoleMarkdownFileUpdateItem struct {
	Name    string `json:"name" validate:"required,oneof=AGENTS.md SOUL.md USER.md IDENTITY.md TOOLS.md HEARTBEAT.md BOOT.md BOOTSTRAP.md"`
	Content string `json:"content"`
}

type AgentRoleMarkdownFilesUpdateReq struct {
	AgentID   uint                              `json:"agentId" validate:"required"`
	Workspace string                            `json:"workspace" validate:"required"`
	Restart   bool                              `json:"restart"`
	Files     []AgentRoleMarkdownFileUpdateItem `json:"files" validate:"required"`
}

type AgentOverviewSnapshot struct {
	ContainerStatus string `json:"containerStatus"`
	AppVersion      string `json:"appVersion"`
	DefaultModel    string `json:"defaultModel"`
	ChannelCount    int    `json:"channelCount"`
	SkillCount      int    `json:"skillCount"`
	JobCount        int    `json:"jobCount"`
	SessionCount    int    `json:"sessionCount"`
}

type AgentAccountModel struct {
	RecordID      uint     `json:"recordId"`
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	ContextWindow int      `json:"contextWindow"`
	MaxTokens     int      `json:"maxTokens"`
	Reasoning     bool     `json:"reasoning"`
	Input         []string `json:"input"`
}

type AgentAccountModelReq struct {
	AccountID uint `json:"accountId" validate:"required"`
}

type AgentAccountModelCreateReq struct {
	AccountID uint              `json:"accountId" validate:"required"`
	Model     AgentAccountModel `json:"model" validate:"required"`
}

type AgentAccountModelUpdateReq struct {
	AccountID uint              `json:"accountId" validate:"required"`
	Model     AgentAccountModel `json:"model" validate:"required"`
}

type AgentAccountModelDeleteReq struct {
	AccountID uint `json:"accountId" validate:"required"`
	RecordID  uint `json:"recordId" validate:"required"`
}

type AgentAccountCreateReq struct {
	Provider       string              `json:"provider" validate:"required"`
	Name           string              `json:"name" validate:"required"`
	APIKey         string              `json:"apiKey" validate:"required"`
	RememberAPIKey bool                `json:"rememberApiKey"`
	BaseURL        string              `json:"baseURL"`
	Models         []AgentAccountModel `json:"models"`
	APIType        string              `json:"apiType" validate:"required"`
	Remark         string              `json:"remark"`
}

type AgentAccountUpdateReq struct {
	ID             uint   `json:"id" validate:"required"`
	Name           string `json:"name" validate:"required"`
	APIKey         string `json:"apiKey" validate:"required"`
	RememberAPIKey bool   `json:"rememberApiKey"`
	BaseURL        string `json:"baseURL"`
	APIType        string `json:"apiType" validate:"required"`
	Remark         string `json:"remark"`
	SyncAgents     bool   `json:"syncAgents"`
}

type AgentAccountVerifyReq struct {
	Provider string `json:"provider" validate:"required"`
	APIKey   string `json:"apiKey" validate:"required"`
	BaseURL  string `json:"baseURL"`
}

type AgentAccountDeleteReq struct {
	ID uint `json:"id" validate:"required"`
}

type AgentAccountSearch struct {
	PageInfo
	Provider string `json:"provider"`
	Name     string `json:"name"`
}

type AgentAccountInfo struct {
	ID             uint                `json:"id"`
	Provider       string              `json:"provider"`
	ProviderName   string              `json:"providerName"`
	Name           string              `json:"name"`
	APIKey         string              `json:"apiKey"`
	RememberAPIKey bool                `json:"rememberApiKey"`
	BaseURL        string              `json:"baseUrl"`
	Models         []AgentAccountModel `json:"models"`
	APIType        string              `json:"apiType"`
	Verified       bool                `json:"verified"`
	Remark         string              `json:"remark"`
	CreatedAt      time.Time           `json:"createdAt"`
}

type ProviderModelInfo struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	ContextWindow int      `json:"contextWindow"`
	MaxTokens     int      `json:"maxTokens"`
	Reasoning     bool     `json:"reasoning"`
	Input         []string `json:"input"`
}

type ProviderInfo struct {
	Sort        uint                `json:"-"`
	Provider    string              `json:"provider"`
	DisplayName string              `json:"displayName"`
	BaseURL     string              `json:"baseUrl"`
	Models      []ProviderModelInfo `json:"models"`
}

type AgentFeishuConfigReq struct {
	AgentID uint `json:"agentId" validate:"required"`
}

type AgentFeishuConfigUpdateReq struct {
	AgentID        uint             `json:"agentId" validate:"required"`
	Enabled        bool             `json:"enabled"`
	ThreadSession  bool             `json:"threadSession"`
	ReplyMode      string           `json:"replyMode" validate:"required"`
	Streaming      bool             `json:"streaming"`
	RequireMention string           `json:"requireMention" validate:"required,oneof=true false open"`
	GroupPolicy    string           `json:"groupPolicy" validate:"required,oneof=open allowlist disabled"`
	GroupAllowFrom []string         `json:"groupAllowFrom"`
	Domain         string           `json:"domain"`
	ConnectionMode string           `json:"connectionMode"`
	Bots           []AgentFeishuBot `json:"bots" validate:"required,min=1"`
}

type AgentFeishuPairingApproveReq struct {
	AgentID     uint   `json:"agentId" validate:"required"`
	PairingCode string `json:"pairingCode" validate:"required"`
}

type AgentFeishuConfig struct {
	Enabled        bool             `json:"enabled"`
	ThreadSession  bool             `json:"threadSession"`
	ReplyMode      string           `json:"replyMode"`
	Streaming      bool             `json:"streaming"`
	RequireMention string           `json:"requireMention"`
	GroupPolicy    string           `json:"groupPolicy"`
	GroupAllowFrom []string         `json:"groupAllowFrom"`
	Domain         string           `json:"domain"`
	ConnectionMode string           `json:"connectionMode"`
	Bots           []AgentFeishuBot `json:"bots"`
	Installed      bool             `json:"installed"`
}

type AgentTelegramConfigReq struct {
	AgentID uint `json:"agentId" validate:"required"`
}

type AgentTelegramConfigUpdateReq struct {
	AgentID        uint               `json:"agentId" validate:"required"`
	Enabled        bool               `json:"enabled"`
	DmPolicy       string             `json:"dmPolicy" validate:"required,oneof=pairing open allowlist disabled"`
	AllowFrom      []string           `json:"allowFrom"`
	RequireMention bool               `json:"requireMention"`
	GroupPolicy    string             `json:"groupPolicy" validate:"required,oneof=open allowlist disabled"`
	GroupAllowFrom []string           `json:"groupAllowFrom"`
	Proxy          string             `json:"proxy"`
	Streaming      string             `json:"streaming" validate:"required,oneof=off partial block progress"`
	DefaultAccount string             `json:"defaultAccount" validate:"required"`
	Bots           []AgentTelegramBot `json:"bots" validate:"required,min=1"`
}

type AgentTelegramConfig struct {
	Enabled        bool               `json:"enabled"`
	DmPolicy       string             `json:"dmPolicy"`
	AllowFrom      []string           `json:"allowFrom"`
	RequireMention bool               `json:"requireMention"`
	GroupPolicy    string             `json:"groupPolicy"`
	GroupAllowFrom []string           `json:"groupAllowFrom"`
	Proxy          string             `json:"proxy"`
	Streaming      string             `json:"streaming"`
	DefaultAccount string             `json:"defaultAccount"`
	Bots           []AgentTelegramBot `json:"bots"`
}

type AgentChannelPairingApproveReq struct {
	AgentID     uint   `json:"agentId" validate:"required"`
	Type        string `json:"type" validate:"required,oneof=feishu telegram discord wecom"`
	PairingCode string `json:"pairingCode" validate:"required"`
	AccountID   string `json:"accountId"`
}

type AgentWecomConfigUpdateReq struct {
	AgentID        uint     `json:"agentId" validate:"required"`
	Enabled        bool     `json:"enabled"`
	DmPolicy       string   `json:"dmPolicy" validate:"required,oneof=pairing open allowlist disabled"`
	AllowFrom      []string `json:"allowFrom"`
	GroupPolicy    string   `json:"groupPolicy" validate:"required,oneof=open allowlist disabled"`
	GroupAllowFrom []string `json:"groupAllowFrom"`
	BotID          string   `json:"botId" validate:"required"`
	Secret         string   `json:"secret" validate:"required"`
}

type AgentWecomConfig struct {
	Enabled        bool     `json:"enabled"`
	DmPolicy       string   `json:"dmPolicy"`
	AllowFrom      []string `json:"allowFrom"`
	GroupPolicy    string   `json:"groupPolicy"`
	GroupAllowFrom []string `json:"groupAllowFrom"`
	BotID          string   `json:"botId"`
	Secret         string   `json:"secret"`
	Installed      bool     `json:"installed"`
}

type AgentDingTalkConfigUpdateReq struct {
	AgentID                         uint               `json:"agentId" validate:"required"`
	Enabled                         bool               `json:"enabled"`
	DmPolicy                        string             `json:"dmPolicy" validate:"required,oneof=allowlist open disabled"`
	AllowFrom                       []string           `json:"allowFrom"`
	GroupPolicy                     string             `json:"groupPolicy" validate:"required,oneof=open allowlist disabled"`
	GroupAllowFrom                  []string           `json:"groupAllowFrom"`
	SeparateSessionByConversation   bool               `json:"separateSessionByConversation"`
	GroupSessionScope               string             `json:"groupSessionScope" validate:"required,oneof=group group_sender"`
	SharedMemoryAcrossConversations bool               `json:"sharedMemoryAcrossConversations"`
	AsyncMode                       bool               `json:"asyncMode"`
	AckText                         string             `json:"ackText"`
	Bots                            []AgentDingTalkBot `json:"bots" validate:"required,min=1"`
}

type AgentDingTalkConfig struct {
	Enabled                         bool               `json:"enabled"`
	DmPolicy                        string             `json:"dmPolicy"`
	AllowFrom                       []string           `json:"allowFrom"`
	GroupPolicy                     string             `json:"groupPolicy"`
	GroupAllowFrom                  []string           `json:"groupAllowFrom"`
	SeparateSessionByConversation   bool               `json:"separateSessionByConversation"`
	GroupSessionScope               string             `json:"groupSessionScope"`
	SharedMemoryAcrossConversations bool               `json:"sharedMemoryAcrossConversations"`
	AsyncMode                       bool               `json:"asyncMode"`
	AckText                         string             `json:"ackText"`
	Bots                            []AgentDingTalkBot `json:"bots"`
	Installed                       bool               `json:"installed"`
}

type AgentWeixinLoginReq struct {
	AgentID uint   `json:"agentId" validate:"required"`
	TaskID  string `json:"taskID" validate:"required"`
}

type AgentQQBotConfigUpdateReq struct {
	AgentID        uint            `json:"agentId" validate:"required"`
	Enabled        bool            `json:"enabled"`
	DmPolicy       string          `json:"dmPolicy"`
	AllowFrom      []string        `json:"allowFrom"`
	GroupPolicy    string          `json:"groupPolicy"`
	GroupAllowFrom []string        `json:"groupAllowFrom"`
	Bots           []AgentQQBotBot `json:"bots" validate:"required,min=1"`
}

type AgentQQBotConfig struct {
	Enabled        bool            `json:"enabled"`
	DmPolicy       string          `json:"dmPolicy"`
	AllowFrom      []string        `json:"allowFrom"`
	GroupPolicy    string          `json:"groupPolicy"`
	GroupAllowFrom []string        `json:"groupAllowFrom"`
	Bots           []AgentQQBotBot `json:"bots"`
	Installed      bool            `json:"installed"`
}

type AgentPluginInstallReq struct {
	AgentID uint   `json:"agentId" validate:"required"`
	Type    string `json:"type" validate:"required,oneof=feishu qqbot wecom dingtalk weixin"`
	TaskID  string `json:"taskID" validate:"required"`
}

type AgentPluginUpgradeReq struct {
	AgentID uint   `json:"agentId" validate:"required"`
	Type    string `json:"type" validate:"required,oneof=feishu qqbot wecom dingtalk weixin"`
	TaskID  string `json:"taskID" validate:"required"`
}

type AgentPluginUninstallReq struct {
	AgentID uint   `json:"agentId" validate:"required"`
	Type    string `json:"type" validate:"required,oneof=feishu qqbot wecom dingtalk weixin"`
	TaskID  string `json:"taskID" validate:"required"`
}

type AgentPluginCheckReq struct {
	AgentID     uint   `json:"agentId" validate:"required"`
	Type        string `json:"type" validate:"required,oneof=feishu qqbot wecom dingtalk weixin"`
	CheckLatest bool   `json:"checkLatest"`
}

type AgentPluginStatus struct {
	Installed      bool   `json:"installed"`
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	Upgradable     bool   `json:"upgradable"`
}

type AgentDiscordConfigUpdateReq struct {
	AgentID        uint              `json:"agentId" validate:"required"`
	Enabled        bool              `json:"enabled"`
	DmPolicy       string            `json:"dmPolicy" validate:"required"`
	AllowFrom      []string          `json:"allowFrom"`
	RequireMention bool              `json:"requireMention"`
	GroupPolicy    string            `json:"groupPolicy" validate:"required,oneof=open allowlist disabled"`
	Proxy          string            `json:"proxy"`
	DefaultAccount string            `json:"defaultAccount" validate:"required"`
	Bots           []AgentDiscordBot `json:"bots" validate:"required,min=1"`
}

type AgentDiscordConfig struct {
	Enabled        bool              `json:"enabled"`
	DmPolicy       string            `json:"dmPolicy"`
	AllowFrom      []string          `json:"allowFrom"`
	RequireMention bool              `json:"requireMention"`
	GroupPolicy    string            `json:"groupPolicy"`
	Proxy          string            `json:"proxy"`
	DefaultAccount string            `json:"defaultAccount"`
	Bots           []AgentDiscordBot `json:"bots"`
}

type AgentChannelBotBase struct {
	AccountID string `json:"accountId"`
	Name      string `json:"name"`
	Enabled   bool   `json:"enabled"`
	IsDefault bool   `json:"isDefault"`
}

func (b AgentChannelBotBase) IsEnabled() bool {
	return b.Enabled
}

type AgentFeishuBot struct {
	AgentChannelBotBase
	AppID     string   `json:"appId"`
	AppSecret string   `json:"appSecret"`
	DmPolicy  string   `json:"dmPolicy"`
	AllowFrom []string `json:"allowFrom"`
}

type AgentTelegramBot struct {
	AgentChannelBotBase
	BotToken    string `json:"botToken"`
	DmPolicy    string `json:"dmPolicy"`
	GroupPolicy string `json:"groupPolicy"`
	Streaming   string `json:"streaming"`
}

type AgentDiscordBot struct {
	AgentChannelBotBase
	Token string `json:"token"`
}

type AgentQQBotBot struct {
	AgentChannelBotBase
	AppID        string   `json:"appId"`
	ClientSecret string   `json:"clientSecret"`
	AllowFrom    []string `json:"allowFrom"`
	SystemPrompt string   `json:"systemPrompt"`
}

type AgentDingTalkBot struct {
	AgentChannelBotBase
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

type AgentSecurityConfigUpdateReq struct {
	AgentID        uint     `json:"agentId" validate:"required"`
	AllowedOrigins []string `json:"allowedOrigins"`
}

type AgentSecurityConfig struct {
	AllowedOrigins []string `json:"allowedOrigins"`
}

type AgentOtherConfigUpdateReq struct {
	AgentID        uint   `json:"agentId" validate:"required"`
	UserTimezone   string `json:"userTimezone" validate:"required"`
	BrowserEnabled bool   `json:"browserEnabled"`
	NPMRegistry    string `json:"npmRegistry" validate:"required"`
}

type AgentOtherConfig struct {
	UserTimezone   string `json:"userTimezone"`
	BrowserEnabled bool   `json:"browserEnabled"`
	NPMRegistry    string `json:"npmRegistry"`
}

type AgentConfigFileReq struct {
	AgentID uint `json:"agentId" validate:"required"`
}

type AgentConfigFileUpdateReq struct {
	AgentID uint   `json:"agentId" validate:"required"`
	Content string `json:"content" validate:"required"`
}

type AgentConfigFile struct {
	Content string `json:"content"`
}

type AgentSkillSearchReq struct {
	AgentID uint   `json:"agentId" validate:"required"`
	Source  string `json:"source" validate:"required,oneof=clawhub-global clawhub-cn skillhub"`
	Keyword string `json:"keyword" validate:"required"`
}

type AgentSkillItem struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Source      string `json:"source"`
	Bundled     bool   `json:"bundled"`
	Disabled    bool   `json:"disabled"`
}

type AgentSkillSearchItem struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Summary     string `json:"summary"`
	Version     string `json:"version"`
	Source      string `json:"source"`
	Score       string `json:"score"`
}

type AgentSkillUpdateReq struct {
	AgentID uint   `json:"agentId" validate:"required"`
	Name    string `json:"name" validate:"required"`
	Enabled bool   `json:"enabled"`
}

type AgentSkillInstallReq struct {
	AgentID uint   `json:"agentId" validate:"required"`
	Source  string `json:"source" validate:"required,oneof=clawhub-global clawhub-cn skillhub"`
	Slug    string `json:"slug" validate:"required"`
	TaskID  string `json:"taskID" validate:"required"`
}
