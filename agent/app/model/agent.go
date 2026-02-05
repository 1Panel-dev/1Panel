package model

type Agent struct {
	BaseModel
	Name         string `json:"name" gorm:"not null;unique"`
	Provider     string `json:"provider"`
	Model        string `json:"model"`
	BaseURL      string `json:"baseUrl"`
	APIKey       string `json:"apiKey"`
	Token        string `json:"token"`
	Status       string `json:"status"`
	Message      string `json:"message"`
	AppInstallID uint   `json:"appInstallId"`
	AccountID    uint   `json:"accountId"`
	ConfigPath   string `json:"configPath"`
}
