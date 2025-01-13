package model

type BackupAccount struct {
	BaseModel
	Name       string `gorm:"not null;default:''" json:"name"`
	Type       string `gorm:"not null;default:''" json:"type"`
	IsPublic   bool   `json:"isPublic"`
	Bucket     string `json:"bucket"`
	AccessKey  string `json:"accessKey"`
	Credential string `json:"credential"`
	BackupPath string `json:"backupPath"`
	Vars       string `json:"vars"`

	RememberAuth bool `json:"rememberAuth"`
}
