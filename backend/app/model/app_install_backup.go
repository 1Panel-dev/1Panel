package model

type AppInstallBackup struct {
	BaseModel
	Name         string `gorm:"type:varchar(64);not null" json:"name"`
	Path         string `gorm:"type:varchar(64);not null" json:"path"`
	Param        string `json:"param"  gorm:"type:longtext;"`
	AppDetailId  uint   `gorm:"type:varchar(64);not null" json:"app_detail_id"`
	AppInstallId uint   `gorm:"type:integer;not null" json:"app_install_id"`
}
