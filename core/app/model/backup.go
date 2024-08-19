package model

import "time"

type BackupAccount struct {
	BaseModel
	Name       string `gorm:"unique;not null" json:"name"`
	Type       string `gorm:"unique;not null" json:"type"`
	Bucket     string `json:"bucket"`
	AccessKey  string `json:"accessKey"`
	Credential string `json:"credential"`
	BackupPath string `json:"backupPath"`
	Vars       string `json:"vars"`

	RememberAuth bool `json:"rememberAuth"`
	EntryID      uint `json:"entryID"`

	DeletedAt time.Time `json:"deletedAt"`
}
