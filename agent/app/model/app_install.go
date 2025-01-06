package model

import (
	"path"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/constant"
)

type AppInstall struct {
	BaseModel
	Name          string `json:"name" gorm:"not null;UNIQUE"`
	AppId         uint   `json:"appId" gorm:"not null"`
	AppDetailId   uint   `json:"appDetailId" gorm:"not null"`
	Version       string `json:"version" gorm:"not null"`
	Param         string `json:"param"`
	Env           string `json:"env"`
	DockerCompose string `json:"dockerCompose" `
	Status        string `json:"status" gorm:"not null"`
	Description   string `json:"description"`
	Message       string `json:"message"`
	ContainerName string `json:"containerName" gorm:"not null"`
	ServiceName   string `json:"serviceName" gorm:"not null"`
	HttpPort      int    `json:"httpPort"`
	HttpsPort     int    `json:"httpsPort"`
	WebUI         string `json:"webUI"`
	App           App    `json:"app" gorm:"-:migration"`
}

func (i *AppInstall) GetPath() string {
	return path.Join(i.GetAppPath(), i.Name)
}

func (i *AppInstall) GetComposePath() string {
	return path.Join(i.GetAppPath(), i.Name, "docker-compose.yml")
}

func (i *AppInstall) GetEnvPath() string {
	return path.Join(i.GetAppPath(), i.Name, ".env")
}

func (i *AppInstall) GetAppPath() string {
	if i.App.Resource == constant.AppResourceLocal {
		return path.Join(constant.LocalAppInstallDir, strings.TrimPrefix(i.App.Key, constant.AppResourceLocal))
	} else {
		return path.Join(constant.AppInstallDir, i.App.Key)
	}
}
