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

type BackupRecord struct {
	BaseModel
	From              string `json:"from"`
	CronjobID         uint   `json:"cronjobID"`
	SourceAccountIDs  string `json:"sourceAccountIDs"`
	DownloadAccountID uint   `json:"downloadAccountID"`

	Type       string `gorm:"not null;default:''" json:"type"`
	Name       string `gorm:"not null;default:''" json:"name"`
	DetailName string `json:"detailName"`
	FileDir    string `json:"fileDir"`
	FileName   string `json:"fileName"`

	TaskID      string `json:"taskID"`
	Status      string `json:"status"`
	Message     string `json:"message"`
	Description string `json:"description"`
}
