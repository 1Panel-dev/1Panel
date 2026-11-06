package model

import "time"

type AgentAccount struct {
	BaseModel
	AppInstallID         uint       `json:"appInstallId"`
	Provider             string     `json:"provider"`
	Name                 string     `json:"name"`
	APIKey               string     `json:"apiKey"`
	BaseURL              string     `json:"baseUrl"`
	APIType              string     `json:"apiType"`
	AuthMode             string     `json:"authMode"`
	VerifyModel          string     `json:"verifyModel"`
	RememberAPIKey       bool       `json:"rememberApiKey"`
	VerifiedAt           *time.Time `json:"verifiedAt"`
	VerifyError          string     `json:"verifyError"`
	Verified             bool       `json:"verified"`
	ValidateAvailability bool       `json:"validateAvailability"`
	Remark               string     `json:"remark"`
	MasterAccountID      uint       `json:"masterAccountId" gorm:"index"`
}

func (AgentAccount) TableName() string {
	return "agent_provider_accounts"
}
