package model

type AgentAccount struct {
	BaseModel
	Provider      string `json:"provider"`
	Name          string `json:"name"`
	APIKey        string `json:"apiKey"`
	BaseURL       string `json:"baseUrl"`
	Model         string `json:"model"`
	APIType       string `json:"apiType"`
	MaxTokens     int    `json:"maxTokens"`
	ContextWindow int    `json:"contextWindow"`
	Verified      bool   `json:"verified"`
	Remark        string `json:"remark"`
}

func (AgentAccount) TableName() string {
	return "agent_provider_accounts"
}
