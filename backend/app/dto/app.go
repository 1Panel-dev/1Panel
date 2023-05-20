package dto

import (
	"github.com/1Panel-dev/1Panel/backend/app/model"
)

type AppDatabase struct {
	ServiceName string `json:"PANEL_DB_HOST"`
	DbName      string `json:"PANEL_DB_NAME"`
	DbUser      string `json:"PANEL_DB_USER"`
	Password    string `json:"PANEL_DB_USER_PASSWORD"`
}

type AuthParam struct {
	RootPassword string `json:"PANEL_DB_ROOT_PASSWORD"`
}

type ContainerExec struct {
	ContainerName string      `json:"containerName"`
	DbParam       AppDatabase `json:"dbParam"`
	Auth          AuthParam   `json:"auth"`
}

type AppOssConfig struct {
	Version string `json:"version"`
	Package string `json:"package"`
}

type AppVersion struct {
	Version  string `json:"version"`
	DetailId uint   `json:"detailId"`
}

type AppList struct {
	Valid        bool     `json:"valid"`
	Violations   []string `json:"violations"`
	LastModified int      `json:"lastModified"`

	Apps  []AppDefine     `json:"apps"`
	Extra ExtraProperties `json:"additionalProperties"`
}

type AppDefine struct {
	Icon         string `json:"icon"`
	Name         string `json:"name"`
	ReadMe       string `json:"readMe"`
	LastModified int    `json:"lastModified"`

	AppProperty AppProperty        `json:"additionalProperties"`
	Versions    []AppConfigVersion `json:"versions"`
}

type LocalAppAppDefine struct {
	AppProperty model.App `json:"additionalProperties" yaml:"additionalProperties"`
}

type LocalAppParam struct {
	AppParams LocalAppInstallDefine `json:"additionalProperties" yaml:"additionalProperties"`
}

type LocalAppInstallDefine struct {
	FormFields interface{} `json:"formFields" yaml:"formFields"`
}

type ExtraProperties struct {
	Tags []Tag `json:"tags"`
}

type AppProperty struct {
	Name               string   `json:"name"`
	Type               string   `json:"type"`
	Tags               []string `json:"tags"`
	ShortDescZh        string   `json:"shortDescZh"`
	ShortDescEn        string   `json:"shortDescEn"`
	Key                string   `json:"key"`
	Required           []string `json:"Required"`
	CrossVersionUpdate bool     `json:"crossVersionUpdate"`
	Limit              int      `json:"limit"`
	Recommend          int      `json:"recommend"`
	Website            string   `json:"website"`
	Github             string   `json:"github"`
	Document           string   `json:"document"`
}

type AppConfigVersion struct {
	Name                string      `json:"name"`
	LastModified        int         `json:"lastModified"`
	DownloadUrl         string      `json:"downloadUrl"`
	DownloadCallBackUrl string      `json:"downloadCallBackUrl"`
	AppForm             interface{} `json:"additionalProperties"`
}

type Tag struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type AppForm struct {
	FormFields []AppFormFields `json:"formFields"`
}

type AppFormFields struct {
	Type     string         `json:"type"`
	LabelZh  string         `json:"labelZh"`
	LabelEn  string         `json:"labelEn"`
	Required bool           `json:"required"`
	Default  interface{}    `json:"default"`
	EnvKey   string         `json:"envKey"`
	Disabled bool           `json:"disabled"`
	Edit     bool           `json:"edit"`
	Rule     string         `json:"rule"`
	Multiple bool           `json:"multiple"`
	Child    interface{}    `json:"child"`
	Values   []AppFormValue `json:"values"`
}

type AppFormValue struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type AppResource struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

var AppToolMap = map[string]string{
	"mysql": "phpmyadmin",
	"redis": "redis-commander",
}
