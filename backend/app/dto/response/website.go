package response

import (
	"github.com/1Panel-dev/1Panel/backend/app/model"
)

type WebsiteDTO struct {
	model.Website
	ErrorLogPath  string `json:"errorLogPath"`
	AccessLogPath string `json:"accessLogPath"`
}

type WebsitePreInstallCheck struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Version string `json:"version"`
	AppName string `json:"appName"`
}

type WebsiteNginxConfig struct {
	Enable bool         `json:"enable"`
	Params []NginxParam `json:"params"`
}

type WebsiteWafConfig struct {
	Enable   bool   `json:"enable"`
	FilePath string `json:"filePath"`
	Content  string `json:"content"`
}

type WebsiteHTTPS struct {
	Enable bool             `json:"enable"`
	SSL    model.WebsiteSSL `json:"SSL"`
}
