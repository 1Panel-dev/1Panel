package request

import (
	"github.com/1Panel-dev/1Panel/backend/app/dto"
)

type WebsiteSearch struct {
	dto.PageInfo
	Name           string `json:"name"`
	OrderBy        string `json:"orderBy" validate:"required,oneof=primary_domain type status created_at expire_date"`
	Order          string `json:"order" validate:"required,oneof=null ascending descending"`
	WebsiteGroupID uint   `json:"websiteGroupId"`
}

type WebsiteCreate struct {
	PrimaryDomain  string `json:"primaryDomain" validate:"required"`
	Type           string `json:"type" validate:"required"`
	Alias          string `json:"alias" validate:"required"`
	Remark         string `json:"remark"`
	OtherDomains   string `json:"otherDomains"`
	Proxy          string `json:"proxy"`
	WebsiteGroupID uint   `json:"webSiteGroupID" validate:"required"`
	IPV6           bool   `json:"IPV6"`

	AppType      string        `json:"appType" validate:"oneof=new installed"`
	AppInstall   NewAppInstall `json:"appInstall"`
	AppID        uint          `json:"appID"`
	AppInstallID uint          `json:"appInstallID"`

	FtpUser     string `json:"ftpUser"`
	FtpPassword string `json:"ftpPassword"`

	RuntimeID uint `json:"runtimeID"`
	RuntimeConfig
}

type RuntimeConfig struct {
	ProxyType string `json:"proxyType"`
	Port      int    `json:"port"`
}

type NewAppInstall struct {
	Name        string                 `json:"name"`
	AppDetailId uint                   `json:"appDetailID"`
	Params      map[string]interface{} `json:"params"`

	AppContainerConfig
}

type WebsiteInstallCheckReq struct {
	InstallIds []uint `json:"InstallIds"`
}

type WebsiteUpdate struct {
	ID             uint   `json:"id" validate:"required"`
	PrimaryDomain  string `json:"primaryDomain" validate:"required"`
	Remark         string `json:"remark"`
	WebsiteGroupID uint   `json:"webSiteGroupID"`
	ExpireDate     string `json:"expireDate"`
	IPV6           bool   `json:"IPV6"`
}

type WebsiteDelete struct {
	ID           uint `json:"id" validate:"required"`
	DeleteApp    bool `json:"deleteApp"`
	DeleteBackup bool `json:"deleteBackup"`
	ForceDelete  bool `json:"forceDelete"`
}

type WebsiteOp struct {
	ID      uint   `json:"id" validate:"required"`
	Operate string `json:"operate"`
}

type WebsiteRedirectUpdate struct {
	WebsiteID uint   `json:"websiteId" validate:"required"`
	Key       string `json:"key" validate:"required"`
	Enable    bool   `json:"enable"`
}

type WebsiteRecover struct {
	WebsiteName string `json:"websiteName" validate:"required"`
	Type        string `json:"type" validate:"required"`
	BackupName  string `json:"backupName" validate:"required"`
}

type WebsiteRecoverByFile struct {
	WebsiteName string `json:"websiteName" validate:"required"`
	Type        string `json:"type" validate:"required"`
	FileDir     string `json:"fileDir" validate:"required"`
	FileName    string `json:"fileName" validate:"required"`
}

type WebsiteGroupCreate struct {
	Name string `json:"name" validate:"required"`
}

type WebsiteGroupUpdate struct {
	ID      uint   `json:"id" validate:"required"`
	Name    string `json:"name"`
	Default bool   `json:"default"`
}

type WebsiteDomainCreate struct {
	WebsiteID uint   `json:"websiteID" validate:"required"`
	Domains   string `json:"domains" validate:"required"`
}

type WebsiteDomainDelete struct {
	ID uint `json:"id" validate:"required"`
}

type WebsiteHTTPSOp struct {
	WebsiteID       uint     `json:"websiteId" validate:"required"`
	Enable          bool     `json:"enable"`
	WebsiteSSLID    uint     `json:"websiteSSLId"`
	Type            string   `json:"type"  validate:"oneof=existed auto manual"`
	PrivateKey      string   `json:"privateKey"`
	Certificate     string   `json:"certificate"`
	PrivateKeyPath  string   `json:"privateKeyPath"`
	CertificatePath string   `json:"certificatePath"`
	ImportType      string   `json:"importType"`
	HttpConfig      string   `json:"httpConfig"  validate:"oneof=HTTPSOnly HTTPAlso HTTPToHTTPS"`
	SSLProtocol     []string `json:"SSLProtocol"`
	Algorithm       string   `json:"algorithm"`
	Hsts            bool     `json:"hsts"`
}

type WebsiteNginxUpdate struct {
	ID      uint   `json:"id" validate:"required"`
	Content string `json:"content" validate:"required"`
}

type WebsiteLogReq struct {
	ID       uint   `json:"id" validate:"required"`
	Operate  string `json:"operate" validate:"required"`
	LogType  string `json:"logType" validate:"required"`
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
}

type WebsiteDefaultUpdate struct {
	ID uint `json:"id"`
}

type WebsitePHPConfigUpdate struct {
	ID               uint              `json:"id" validate:"required"`
	Params           map[string]string `json:"params"`
	Scope            string            `json:"scope" validate:"required"`
	DisableFunctions []string          `json:"disableFunctions"`
	UploadMaxSize    string            `json:"uploadMaxSize"`
}

type WebsitePHPFileUpdate struct {
	ID      uint   `json:"id" validate:"required"`
	Type    string `json:"type" validate:"required"`
	Content string `json:"content" validate:"required"`
}

type WebsitePHPVersionReq struct {
	WebsiteID    uint `json:"websiteID" validate:"required"`
	RuntimeID    uint `json:"runtimeID" validate:"required"`
	RetainConfig bool `json:"retainConfig" `
}

type WebsiteUpdateDir struct {
	ID      uint   `json:"id" validate:"required"`
	SiteDir string `json:"siteDir" validate:"required"`
}

type WebsiteUpdateDirPermission struct {
	ID    uint   `json:"id" validate:"required"`
	User  string `json:"user" validate:"required"`
	Group string `json:"group" validate:"required"`
}

type WebsiteProxyConfig struct {
	ID           uint              `json:"id" validate:"required"`
	Operate      string            `json:"operate" validate:"required"`
	Enable       bool              `json:"enable" `
	Cache        bool              `json:"cache" `
	CacheTime    int               `json:"cacheTime"  `
	CacheUnit    string            `json:"cacheUnit"`
	Name         string            `json:"name" validate:"required"`
	Modifier     string            `json:"modifier"`
	Match        string            `json:"match" validate:"required"`
	ProxyPass    string            `json:"proxyPass" validate:"required"`
	ProxyHost    string            `json:"proxyHost" validate:"required"`
	Content      string            `json:"content"`
	FilePath     string            `json:"filePath"`
	Replaces     map[string]string `json:"replaces"`
	SNI          bool              `json:"sni"`
	ProxySSLName string            `json:"proxySSLName"`
}

type WebsiteProxyDel struct {
	ID   uint   `json:"id" validate:"required"`
	Name string `json:"name" validate:"required"`
}

type WebsiteProxyReq struct {
	ID uint `json:"id" validate:"required"`
}

type WebsiteRedirectReq struct {
	WebsiteID uint `json:"websiteId" validate:"required"`
}

type WebsiteCommonReq struct {
	ID uint `json:"id" validate:"required"`
}

type WafWebsite struct {
	Key     string   `json:"key"`
	Domains []string `json:"domains"`
	Host    []string `json:"host"`
}

type WebsiteHtmlReq struct {
	Type string `json:"type" validate:"required"`
}

type WebsiteHtmlUpdate struct {
	Type    string `json:"type" validate:"required"`
	Content string `json:"content" validate:"required"`
}
