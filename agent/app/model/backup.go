package model

type BackupRecord struct {
	BaseModel
	From              string `json:"from"`
	CronjobID         uint   `json:"cronjobID"`
	SourceAccountIDs  string `json:"sourceAccountsIDs"`
	DownloadAccountID uint   `json:"downloadAccountID"`

	Type       string `gorm:"not null" json:"type"`
	Name       string `gorm:"not null" json:"name"`
	DetailName string `json:"detailName"`
	FileDir    string `json:"fileDir"`
	FileName   string `json:"fileName"`
}
