package response

import (
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/model"
)

type WebsiteDTO struct {
	model.Website
	ErrorLogPath  string `json:"errorLogPath"`
	AccessLogPath string `json:"accessLogPath"`
	SitePath      string `json:"sitePath"`
	AppName       string `json:"appName"`
	RuntimeName   string `json:"runtimeName"`
	RuntimeType   string `json:"runtimeType"`
	SiteDir       string `json:"siteDir"`
	OpenBaseDir   bool   `json:"openBaseDir"`
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
	ChildSites    []string  `json:"childSites"`
	RuntimeType   string    `json:"runtimeType"`
	Favorite      bool      `json:"favorite"`
	IPV6          bool      `json:"IPV6"`
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
	HttpsPorts  []int            `json:"httpsPorts"`
	HttpsPort   string           `json:"httpsPort"`
	Http3       bool             `json:"http3"`
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

type WebsiteRealIP struct {
	WebsiteID uint   `json:"websiteID" validate:"required"`
	Open      bool   `json:"open"`
	IPFrom    string `json:"ipFrom"`
	IPHeader  string `json:"ipHeader"`
	IPOther   string `json:"ipOther"`
}

type Resource struct {
	Name       string      `json:"name"`
	Type       string      `json:"type"`
	ResourceID uint        `json:"resourceID"`
	Detail     interface{} `json:"detail"`
}

type Database struct {
	Name         string `json:"name"`
	DatabaseName string `json:"databaseName"`
	Type         string `json:"type"`
	From         string `json:"from"`
	ID           uint   `json:"id"`
}
