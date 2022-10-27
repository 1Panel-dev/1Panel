package model

type BackupAccount struct {
	BaseModel
	Type       string `gorm:"type:varchar(64);unique;not null" json:"type"`
	Bucket     string `gorm:"type:varchar(256)" json:"bucket"`
	Credential string `gorm:"type:varchar(256)" json:"credential"`
	Vars       string `gorm:"type:longText" json:"vars"`
}

type BackupRecord struct {
	BaseModel
	Type       string `gorm:"type:varchar(64);not null" json:"type"`
	Name       string `gorm:"type:varchar(64);not null" json:"name"`
	DetailName string `gorm:"type:varchar(256)" json:"detailName"`
	Source     string `gorm:"type:varchar(256)" json:"source"`
	FileDir    string `gorm:"type:varchar(256)" json:"fileDir"`
	FileName   string `gorm:"type:varchar(256)" json:"fileName"`
}
