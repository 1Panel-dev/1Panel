package model

type AppContainer struct {
	BaseModel
	ServiceName   string     `json:"serviceName" gorm:"type:varchar(64);not null"`
	ContainerName string     `json:"containerName" gorm:"type:varchar(64);not null"`
	AppInstallID  uint       `json:"appInstallId" gorm:"type:integer;not null"`
	Port          int        `json:"port" gorm:"type:integer;not null"`
	Auth          string     `json:"auth" gorm:"type:longtext;not null"`
	AppInstall    AppInstall `gorm:"-"`
}
