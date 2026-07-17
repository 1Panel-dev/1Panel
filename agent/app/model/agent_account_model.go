package model

type AgentAccountModel struct {
	BaseModel
	AccountID uint   `json:"accountId" gorm:"index"`
	Model     string `json:"model" gorm:"index"`
	Name      string `json:"name"`
	SortOrder int    `json:"sortOrder" gorm:"index"`
}

func (AgentAccountModel) TableName() string {
	return "agent_account_models"
}
