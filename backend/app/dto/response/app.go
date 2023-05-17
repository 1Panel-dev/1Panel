package response

import (
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/model"
)

type AppRes struct {
	Items []*AppDTO `json:"items"`
	Total int64     `json:"total"`
}

type AppUpdateRes struct {
	CanUpdate            bool        `json:"canUpdate"`
	AppStoreLastModified int         `json:"appStoreLastModified"`
	List                 dto.AppList `json:"list"`
}

type AppDTO struct {
	model.App
	Versions []string    `json:"versions"`
	Tags     []model.Tag `json:"tags"`
}

type TagDTO struct {
	model.Tag
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
	Image  string      `json:"image"`
}

type AppInstalledDTO struct {
	model.AppInstall
	Total     int    `json:"total"`
	Ready     int    `json:"ready"`
	AppName   string `json:"appName"`
	Icon      string `json:"icon"`
	CanUpdate bool   `json:"canUpdate"`
}

type DatabaseConn struct {
	Password    string `json:"password"`
	ServiceName string `json:"serviceName"`
	Port        int64  `json:"port"`
}

type AppService struct {
	Label  string      `json:"label"`
	Value  string      `json:"value"`
	Config interface{} `json:"config"`
}

type AppParam struct {
	Value     interface{} `json:"value"`
	Edit      bool        `json:"edit"`
	Key       string      `json:"key"`
	Rule      string      `json:"rule"`
	LabelZh   string      `json:"labelZh"`
	LabelEn   string      `json:"labelEn"`
	Type      string      `json:"type"`
	Values    interface{} `json:"values"`
	ShowValue string      `json:"showValue"`
	Required  bool        `json:"required"`
	Multiple  bool        `json:"multiple"`
}

type AppConfig struct {
	Params        []AppParam `json:"params"`
	CpuQuota      float64    `json:"cpuQuota"`
	MemoryLimit   float64    `json:"memoryLimit"`
	MemoryUnit    string     `json:"memoryUnit"`
	ContainerName string     `json:"containerName"`
	AllowPort     bool       `json:"allowPort"`
}
