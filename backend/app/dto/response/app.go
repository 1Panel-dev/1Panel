package response

import (
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/dto/request"

	"github.com/1Panel-dev/1Panel/backend/app/model"
)

type AppRes struct {
	Items []*AppItem `json:"items"`
	Total int64      `json:"total"`
}

type AppUpdateRes struct {
	CanUpdate            bool         `json:"canUpdate"`
	IsSyncing            bool         `json:"isSyncing"`
	AppStoreLastModified int          `json:"appStoreLastModified"`
	AppList              *dto.AppList `json:"appList"`
}

type AppDTO struct {
	model.App
	Installed bool     `json:"installed"`
	Versions  []string `json:"versions"`
	Tags      []TagDTO `json:"tags"`
}

type AppItem struct {
	Name        string   `json:"name"`
	Key         string   `json:"key"`
	ID          uint     `json:"id"`
	Description string   `json:"description"`
	Icon        string   `json:"icon"`
	Type        string   `json:"type"`
	Status      string   `json:"status"`
	Resource    string   `json:"resource"`
	Installed   bool     `json:"installed"`
	Versions    []string `json:"versions"`
	Limit       int      `json:"limit"`
	Tags        []TagDTO `json:"tags"`
	GpuSupport  bool     `json:"gpuSupport"`
}

type TagDTO struct {
	ID   uint   `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
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
	HttpPort      int       `json:"httpPort"`
	HttpsPort     int       `json:"httpsPort"`
}

type AppDetailDTO struct {
	model.AppDetail
	Enable     bool        `json:"enable"`
	Params     interface{} `json:"params"`
	Image      string      `json:"image"`
	HostMode   bool        `json:"hostMode"`
	GpuSupport bool        `json:"gpuSupport"`
}

type IgnoredApp struct {
	Icon     string `json:"icon"`
	Name     string `json:"name"`
	Version  string `json:"version"`
	DetailID uint   `json:"detailID"`
}

type AppInstalledDTO struct {
	model.AppInstall
	Total     int    `json:"total"`
	Ready     int    `json:"ready"`
	AppName   string `json:"appName"`
	Icon      string `json:"icon"`
	CanUpdate bool   `json:"canUpdate"`
	Path      string `json:"path"`
}

type AppDetail struct {
	Website  string `json:"website"`
	Document string `json:"document"`
	Github   string `json:"github"`
}

type AppInstallDTO struct {
	ID            uint      `json:"id"`
	Name          string    `json:"name"`
	AppID         uint      `json:"appID"`
	AppDetailID   uint      `json:"appDetailID"`
	Version       string    `json:"version"`
	Status        string    `json:"status"`
	Message       string    `json:"message"`
	HttpPort      int       `json:"httpPort"`
	HttpsPort     int       `json:"httpsPort"`
	Path          string    `json:"path"`
	CanUpdate     bool      `json:"canUpdate"`
	Icon          string    `json:"icon"`
	AppName       string    `json:"appName"`
	Ready         int       `json:"ready"`
	Total         int       `json:"total"`
	AppKey        string    `json:"appKey"`
	AppType       string    `json:"appType"`
	AppStatus     string    `json:"appStatus"`
	DockerCompose string    `json:"dockerCompose"`
	CreatedAt     time.Time `json:"createdAt"`
	App           AppDetail `json:"app"`
}

type DatabaseConn struct {
	Status        string `json:"status"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	ContainerName string `json:"containerName"`
	ServiceName   string `json:"serviceName"`
	Port          int64  `json:"port"`
}

type AppService struct {
	Label  string      `json:"label"`
	Value  string      `json:"value"`
	Config interface{} `json:"config"`
	From   string      `json:"from"`
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
	Params []AppParam `json:"params"`
	request.AppContainerConfig
}
