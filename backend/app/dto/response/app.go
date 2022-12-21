package response

import (
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"time"
)

type AppRes struct {
	Version   string      `json:"version"`
	CanUpdate bool        `json:"canUpdate"`
	Items     []*AppDTO   `json:"items"`
	Tags      []model.Tag `json:"tags"`
	Total     int64       `json:"total"`
}

type AppDTO struct {
	model.App
	Versions []string    `json:"versions"`
	Tags     []model.Tag `json:"tags"`
}

type AppInstalledCheck struct {
	IsExist       bool      `json:"isExist"`
	Name          string    `json:"name"`
	App           string    `json:"app"`
	Version       string    `json:"version"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
	LastBackupAt  string    `json:"lastBackupAt"`
	AppInstallID  uint      `json:"appInstallId"`
	ContainerName string    `json:"containerName"`
	InstallPath   string    `json:"installPath"`
}

type AppDetailDTO struct {
	model.AppDetail
	Enable bool        `json:"enable"`
	Params interface{} `json:"params"`
}

type AppInstalledDTO struct {
	model.AppInstall
	Total     int    `json:"total"`
	Ready     int    `json:"ready"`
	AppName   string `json:"appName"`
	Icon      string `json:"icon"`
	CanUpdate bool   `json:"canUpdate"`
}

type AppService struct {
	Label  string      `json:"label"`
	Value  string      `json:"value"`
	Config interface{} `json:"config"`
}
