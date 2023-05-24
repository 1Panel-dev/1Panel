package response

import (
	"github.com/1Panel-dev/1Panel/backend/app/model"
)

type WebsiteDTO struct {
	model.Website
	ErrorLogPath  string `json:"errorLogPath"`
	AccessLogPath string `json:"accessLogPath"`
	SitePath      string `json:"sitePath"`
	AppName       string `json:"appName"`
	RuntimeName   string `json:"runtimeName"`
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
	Enable      bool             `json:"enable"`
	HttpConfig  string           `json:"httpConfig"`
	SSL         model.WebsiteSSL `json:"SSL"`
	SSLProtocol []string         `json:"SSLProtocol"`
	Algorithm   string           `json:"algorithm"`
}

type WebsiteLog struct {
	Enable  bool   `json:"enable"`
	Content string `json:"content"`
}

type PHPConfig struct {
	Params           map[string]string `json:"params"`
	DisableFunctions []string          `json:"disableFunctions"`
	UploadMaxSize    string            `json:"uploadMaxSize"`
}

type NginxRewriteRes struct {
	Content string `json:"content"`
}
