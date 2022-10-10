package model

type Database struct {
	BaseModel
	AppContainerId uint   `json:"appContainerId" gorm:"type:integer;not null"`
	AppInstallId   uint   `json:"appInstallId" gorm:"type:integer;not null"`
	Key            string `json:"key" gorm:"type:varchar(64);not null"`
	Dbname         string `json:"dbname" gorm:"type:varchar(256);not null"`
	Username       string `json:"username" gorm:"type:varchar(256);not null"`
	Password       string `json:"password" gorm:"type:varchar(256);not null"`
}
