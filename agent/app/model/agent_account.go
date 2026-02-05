package model

type AgentAccount struct {
	BaseModel
	Provider string `json:"provider"`
	Name     string `json:"name"`
	APIKey   string `json:"apiKey"`
	BaseURL  string `json:"baseUrl"`
	Verified bool   `json:"verified"`
	Remark   string `json:"remark"`
}

func (AgentAccount) TableName() string {
	return "agent_provider_accounts"
}
