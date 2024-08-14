package model

type BackupAccount struct {
	BaseModel
	Name       string `json:"name"`
	Type       string `json:"type"`
	Bucket     string `json:"bucket"`
	AccessKey  string `json:"accessKey"`
	Credential string `json:"credential"`
	BackupPath string `json:"backupPath"`
	Vars       string `json:"vars"`
}

// Source ---> SourceAccountIDs
// BackupType ---> DownloadAccountID
type BackupRecord struct {
	BaseModel
	From              string `json:"from"`
	CronjobID         uint   `json:"cronjobID"`
	SourceAccountIDs  string `json:"sourceAccountIDs"`
	DownloadAccountID uint   `json:"downloadAccountID"`

	Type       string `gorm:"not null" json:"type"`
	Name       string `gorm:"not null" json:"name"`
	DetailName string `json:"detailName"`
	FileDir    string `json:"fileDir"`
	FileName   string `json:"fileName"`
}
