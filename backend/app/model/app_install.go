package model

type AppInstall struct {
	BaseModel
	Name          string `json:"name" gorm:"type:varchar(64);not null"`
	ContainerName string `json:"containerName" gorm:"type:varchar(256);not null"`
	Version       string `json:"version" gorm:"type:varchar(256);not null"`
	AppId         uint   `json:"appId" gorm:"type:integer;not null"`
	AppDetailId   uint   `json:"appDetailId" gorm:"type:integer;not null"`
	Params        string `json:"params"  gorm:"type:longtext;not null"`
	Status        string `json:"status" gorm:"type:varchar(256);not null"`
	Description   string `json:"description" gorm:"type:varchar(256);not null"`
	Message       string `json:"message"  gorm:"type:longtext;not null"`
	App           App    `json:"-"`
}
