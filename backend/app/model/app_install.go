package model

import (
	"github.com/1Panel-dev/1Panel/backend/constant"
	"path"
)

type AppInstall struct {
	BaseModel
	Name          string             `json:"name" gorm:"type:varchar(64);not null;UNIQUE"`
	AppId         uint               `json:"appId" gorm:"type:integer;not null"`
	AppDetailId   uint               `json:"appDetailId" gorm:"type:integer;not null"`
	Version       string             `json:"version" gorm:"type:varchar(64);not null"`
	Param         string             `json:"param"  gorm:"type:longtext;"`
	Env           string             `json:"env"  gorm:"type:longtext;"`
	DockerCompose string             `json:"dockerCompose"  gorm:"type:longtext;"`
	Status        string             `json:"status" gorm:"type:varchar(256);not null"`
	Description   string             `json:"description" gorm:"type:varchar(256);"`
	Message       string             `json:"message"  gorm:"type:longtext;"`
	ContainerName string             `json:"containerName" gorm:"type:varchar(256);not null"`
	ServiceName   string             `json:"serviceName" gorm:"type:varchar(256);not null"`
	HttpPort      int                `json:"httpPort" gorm:"type:integer;not null"`
	HttpsPort     int                `json:"httpsPort" gorm:"type:integer;not null"`
	App           App                `json:"app" gorm:"-:migration"`
	Backups       []AppInstallBackup `json:"backups" gorm:"-:migration"`
}

func (i *AppInstall) GetPath() string {
	return path.Join(constant.AppInstallDir, i.App.Key, i.Name)
}

func (i *AppInstall) GetComposePath() string {
	return path.Join(constant.AppInstallDir, i.App.Key, i.Name, "docker-compose.yml")
}
