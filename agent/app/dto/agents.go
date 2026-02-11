package dto

import "time"

type AgentCreateReq struct {
	Name          string  `json:"name" validate:"required"`
	AppVersion    string  `json:"appVersion" validate:"required"`
	WebUIPort     int     `json:"webUIPort" validate:"required"`
	BridgePort    int     `json:"bridgePort" validate:"required"`
	Provider      string  `json:"provider" validate:"required"`
	Model         string  `json:"model" validate:"required"`
	AccountID     uint    `json:"accountId"`
	APIKey        string  `json:"apiKey"`
	BaseURL       string  `json:"baseURL"`
	Token         string  `json:"token"`
	TaskID        string  `json:"taskID"`
	Advanced      bool    `json:"advanced"`
	ContainerName string  `json:"containerName"`
	AllowPort     bool    `json:"allowPort"`
	SpecifyIP     string  `json:"specifyIP"`
	RestartPolicy string  `json:"restartPolicy"`
	CpuQuota      float64 `json:"cpuQuota"`
	MemoryLimit   float64 `json:"memoryLimit"`
	MemoryUnit    string  `json:"memoryUnit"`
	PullImage     bool    `json:"pullImage"`
	EditCompose   bool    `json:"editCompose"`
	DockerCompose string  `json:"dockerCompose"`
}

type AgentItem struct {
	ID           uint      `json:"id"`
	Name         string    `json:"name"`
	Provider     string    `json:"provider"`
	ProviderName string    `json:"providerName"`
	Model        string    `json:"model"`
	BaseURL      string    `json:"baseUrl"`
	APIKey       string    `json:"apiKey"`
	Token        string    `json:"token"`
	Status       string    `json:"status"`
	Message      string    `json:"message"`
	AppInstallID uint      `json:"appInstallId"`
	AccountID    uint      `json:"accountId"`
	AppVersion   string    `json:"appVersion"`
	Container    string    `json:"containerName"`
	WebUIPort    int       `json:"webUIPort"`
	BridgePort   int       `json:"bridgePort"`
	Path         string    `json:"path"`
	ConfigPath   string    `json:"configPath"`
	Upgradable   bool      `json:"upgradable"`
	CreatedAt    time.Time `json:"createdAt"`
}

type AgentDeleteReq struct {
	ID          uint   `json:"id" validate:"required"`
	TaskID      string `json:"taskID"`
	ForceDelete bool   `json:"forceDelete"`
}

type AgentTokenResetReq struct {
	ID uint `json:"id" validate:"required"`
}

type AgentModelConfigUpdateReq struct {
	AgentID   uint   `json:"agentId" validate:"required"`
	AccountID uint   `json:"accountId" validate:"required"`
	Model     string `json:"model" validate:"required"`
}

type AgentAccountCreateReq struct {
	Provider string `json:"provider" validate:"required"`
	Name     string `json:"name" validate:"required"`
	APIKey   string `json:"apiKey" validate:"required"`
	BaseURL  string `json:"baseURL"`
	Remark   string `json:"remark"`
}

type AgentAccountUpdateReq struct {
	ID         uint   `json:"id" validate:"required"`
	Name       string `json:"name" validate:"required"`
	APIKey     string `json:"apiKey" validate:"required"`
	BaseURL    string `json:"baseURL"`
	Remark     string `json:"remark"`
	SyncAgents bool   `json:"syncAgents"`
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
	ID           uint      `json:"id"`
	Provider     string    `json:"provider"`
	ProviderName string    `json:"providerName"`
	Name         string    `json:"name"`
	APIKey       string    `json:"apiKey"`
	BaseURL      string    `json:"baseUrl"`
	Verified     bool      `json:"verified"`
	Remark       string    `json:"remark"`
	CreatedAt    time.Time `json:"createdAt"`
}

type ProviderModelInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
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
	AgentID   uint   `json:"agentId" validate:"required"`
	BotName   string `json:"botName" validate:"required"`
	AppID     string `json:"appId" validate:"required"`
	AppSecret string `json:"appSecret" validate:"required"`
	Enabled   bool   `json:"enabled"`
	DmPolicy  string `json:"dmPolicy" validate:"required"`
}

type AgentFeishuPairingApproveReq struct {
	AgentID     uint   `json:"agentId" validate:"required"`
	PairingCode string `json:"pairingCode" validate:"required"`
}

type AgentFeishuConfig struct {
	Enabled   bool   `json:"enabled"`
	DmPolicy  string `json:"dmPolicy"`
	BotName   string `json:"botName"`
	AppID     string `json:"appId"`
	AppSecret string `json:"appSecret"`
}
