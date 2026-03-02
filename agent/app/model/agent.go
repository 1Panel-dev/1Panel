package model

type Agent struct {
	BaseModel
	Name          string `json:"name" gorm:"not null;unique"`
	AgentType     string `json:"agentType" gorm:"default:openclaw"`
	Provider      string `json:"provider"`
	Model         string `json:"model"`
	APIType       string `json:"apiType"`
	MaxTokens     int    `json:"maxTokens"`
	ContextWindow int    `json:"contextWindow"`
	BaseURL       string `json:"baseUrl"`
	APIKey        string `json:"apiKey"`
	Token         string `json:"token"`
	Status        string `json:"status"`
	Message       string `json:"message"`
	AppInstallID  uint   `json:"appInstallId"`
	AccountID     uint   `json:"accountId"`
	ConfigPath    string `json:"configPath"`
}
