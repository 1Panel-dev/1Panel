package model

type BackupAccount struct {
	BaseModel
	Name       string `gorm:"type:varchar(64);not null" json:"name"`
	Type       string `gorm:"type:varchar(64)" json:"type"`
	Bucket     string `gorm:"type:varchar(256)" json:"bucket"`
	Credential string `gorm:"type:varchar(256)" json:"credential"`
	Vars       string `gorm:"type:longText" json:"vars"`
	Status     string `gorm:"type:varchar(64)" json:"status"`
}
