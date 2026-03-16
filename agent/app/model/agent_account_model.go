package model

type AgentAccountModel struct {
	BaseModel
	AccountID     uint   `json:"accountId" gorm:"index"`
	Model         string `json:"model" gorm:"index"`
	Name          string `json:"name"`
	ContextWindow int    `json:"contextWindow"`
	MaxTokens     int    `json:"maxTokens"`
	Reasoning     bool   `json:"reasoning"`
	Input         string `json:"input" gorm:"type:text"`
	SortOrder     int    `json:"sortOrder" gorm:"index"`
}

func (AgentAccountModel) TableName() string {
	return "agent_account_models"
}
