package model

type Snapshot struct {
	BaseModel
	Name        string `json:"name" gorm:"type:varchar(64);not null;unique"`
	Description string `json:"description" gorm:"type:varchar(256)"`
	BackupType  string `json:"backupType" gorm:"type:varchar(64)"`
	Status      string `json:"status" gorm:"type:varchar(64)"`
	Message     string `json:"message" gorm:"type:varchar(256)"`
	Version     string `json:"version" gorm:"type:varchar(256)"`
}
