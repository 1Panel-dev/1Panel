package response

import (
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"time"
)

type WebsiteDTO struct {
	model.Website
	ErrorLogPath  string `json:"errorLogPath"`
	AccessLogPath string `json:"accessLogPath"`
	SitePath      string `json:"sitePath"`
	AppName       string `json:"appName"`
	RuntimeName   string `json:"runtimeName"`
	SiteDir       string `gorm:"type:varchar;" json:"siteDir"`
}

type WebsiteRes struct {
	ID            uint      `json:"id"`
	CreatedAt     time.Time `json:"createdAt"`
	Protocol      string    `json:"protocol"`
	PrimaryDomain string    `json:"primaryDomain"`
	Type          string    `json:"type"`
	Alias         string    `json:"alias"`
	Remark        string    `json:"remark"`
	Status        string    `json:"status"`
	ExpireDate    time.Time `json:"expireDate"`
	SitePath      string    `json:"sitePath"`
	AppName       string    `json:"appName"`
	RuntimeName   string    `json:"runtimeName"`
	SSLExpireDate time.Time `json:"sslExpireDate"`
	SSLStatus     string    `json:"sslStatus"`
	AppInstallID  uint      `json:"appInstallId"`
	RuntimeType   string    `json:"runtimeType"`
}

type WebsiteOption struct {
	ID            uint   `json:"id"`
	PrimaryDomain string `json:"primaryDomain"`
	Alias         string `json:"alias"`
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

type WebsiteHTTPS struct {
	Enable      bool             `json:"enable"`
	HttpConfig  string           `json:"httpConfig"`
	SSL         model.WebsiteSSL `json:"SSL"`
	SSLProtocol []string         `json:"SSLProtocol"`
	Algorithm   string           `json:"algorithm"`
	Hsts        bool             `json:"hsts"`
}

type WebsiteLog struct {
	Enable  bool   `json:"enable"`
	Content string `json:"content"`
	End     bool   `json:"end"`
	Path    string `json:"path"`
}

type PHPConfig struct {
	Params           map[string]string `json:"params"`
	DisableFunctions []string          `json:"disableFunctions"`
	UploadMaxSize    string            `json:"uploadMaxSize"`
}

type NginxRewriteRes struct {
	Content string `json:"content"`
}

type WebsiteDirConfig struct {
	Dirs      []string `json:"dirs"`
	User      string   `json:"user"`
	UserGroup string   `json:"userGroup"`
	Msg       string   `json:"msg"`
}

type WebsiteHtmlRes struct {
	Content string `json:"content"`
}
