package model

type BackupAccount struct {
	BaseModel
	Type       string `gorm:"type:varchar(64);unique;not null" json:"type"`
	Bucket     string `gorm:"type:varchar(256)" json:"bucket"`
	Credential string `gorm:"type:varchar(256)" json:"credential"`
	Vars       string `gorm:"type:longText" json:"vars"`
}
