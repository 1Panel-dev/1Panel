package model

type BackupAccount struct {
	BaseModel
	Name       string `gorm:"type:varchar(64);unique;not null" json:"name"`
	Type       string `gorm:"type:varchar(64);unique;not null" json:"type"`
	Bucket     string `gorm:"type:varchar(256)" json:"bucket"`
	AccessKey  string `gorm:"type:varchar(256)" json:"accessKey"`
	Credential string `gorm:"type:varchar(256)" json:"credential"`
	BackupPath string `gorm:"type:varchar(256)" json:"backupPath"`
	Vars       string `gorm:"type:longText" json:"vars"`

	RememberAuth bool `gorm:"type:varchar(64)" json:"rememberAuth"`
	InUsed       bool `gorm:"type:varchar(64)" json:"inUsed"`
	EntryID      uint `gorm:"type:varchar(64)" json:"entryID"`
}
