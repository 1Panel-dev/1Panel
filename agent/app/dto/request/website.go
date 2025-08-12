package request

import (
	"github.com/1Panel-dev/1Panel/agent/app/dto"
)

type WebsiteSearch struct {
	dto.PageInfo
	Name           string `json:"name"`
	OrderBy        string `json:"orderBy" validate:"required,oneof=primary_domain type status createdAt expire_date created_at favorite"`
	Order          string `json:"order" validate:"required,oneof=null ascending descending"`
	WebsiteGroupID uint   `json:"websiteGroupId"`
	Type           string `json:"type"`
}

type WebsiteCreate struct {
	Type           string `json:"type" validate:"required"`
	Alias          string `json:"alias" validate:"required"`
	Remark         string `json:"remark"`
	Proxy          string `json:"proxy"`
	WebsiteGroupID uint   `json:"webSiteGroupID" validate:"required"`
	IPV6           bool   `json:"IPV6"`

	Domains []WebsiteDomain `json:"domains"`

	AppType      string        `json:"appType" validate:"oneof=new installed"`
	AppInstall   NewAppInstall `json:"appInstall"`
	AppID        uint          `json:"appID"`
	AppInstallID uint          `json:"appInstallID"`

	RuntimeID       uint   `json:"runtimeID"`
	TaskID          string `json:"taskID"`
	ParentWebsiteID uint   `json:"parentWebsiteID"`

	SiteDir string `json:"siteDir"`

	RuntimeConfig
	FtpConfig
	DataBaseConfig
	SSLConfig
}

type WebsiteOptionReq struct {
	Types []string `json:"types"`
}

type RuntimeConfig struct {
	ProxyType string `json:"proxyType"`
	Port      int    `json:"port"`
}

type FtpConfig struct {
	FtpUser     string `json:"ftpUser"`
	FtpPassword string `json:"ftpPassword"`
}

type DataBaseConfig struct {
	CreateDb   bool   `json:"createDb"`
	DbName     string `json:"dbName"`
	DbUser     string `json:"dbUser"`
	DbPassword string `json:"dbPassword"`
	DbHost     string `json:"dbHost"`
	DBFormat   string `json:"dbFormat"`
}

type SSLConfig struct {
	EnableSSL    bool `json:"enableSSL"`
	WebsiteSSLID uint `json:"websiteSSLID"`
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
	Favorite       bool   `json:"favorite"`
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
	WebsiteID uint            `json:"websiteID" validate:"required"`
	Domains   []WebsiteDomain `json:"domains" validate:"required"`
}

type WebsiteDomainUpdate struct {
	ID  uint `json:"id" validate:"required"`
	SSL bool `json:"ssl"`
}

type WebsiteDomain struct {
	Domain string `json:"domain" validate:"required"`
	Port   int    `json:"port"`
	SSL    bool   `json:"ssl"`
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
	HstsIncludeSubDomains            bool     `json:"hstsIncludeSubDomains"`
	HttpsPorts      []int    `json:"httpsPorts"`
	Http3           bool     `json:"http3"`
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

type WebsitePHPVersionReq struct {
	WebsiteID uint `json:"websiteID" validate:"required"`
	RuntimeID uint `json:"runtimeID"`
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
	ID              uint              `json:"id" validate:"required"`
	Operate         string            `json:"operate" validate:"required"`
	Enable          bool              `json:"enable" `
	Cache           bool              `json:"cache" `
	CacheTime       int               `json:"cacheTime"`
	CacheUnit       string            `json:"cacheUnit"`
	ServerCacheTime int               `json:"serverCacheTime"`
	ServerCacheUnit string            `json:"serverCacheUnit"`
	Name            string            `json:"name" validate:"required"`
	Modifier        string            `json:"modifier"`
	Match           string            `json:"match" validate:"required"`
	ProxyPass       string            `json:"proxyPass" validate:"required"`
	ProxyHost       string            `json:"proxyHost" validate:"required"`
	Content         string            `json:"content"`
	FilePath        string            `json:"filePath"`
	Replaces        map[string]string `json:"replaces"`
	SNI             bool              `json:"sni"`
	ProxySSLName    string            `json:"proxySSLName"`
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

type WebsiteLBCreate struct {
	WebsiteID uint                      `json:"websiteID" validate:"required"`
	Name      string                    `json:"name" validate:"required"`
	Algorithm string                    `json:"algorithm"`
	Servers   []dto.NginxUpstreamServer `json:"servers"`
}

type WebsiteLBUpdate struct {
	WebsiteID uint                      `json:"websiteID" validate:"required"`
	Name      string                    `json:"name" validate:"required"`
	Algorithm string                    `json:"algorithm"`
	Servers   []dto.NginxUpstreamServer `json:"servers"`
}

type WebsiteLBDelete struct {
	WebsiteID uint   `json:"websiteID" validate:"required"`
	Name      string `json:"name" validate:"required"`
}

type WebsiteLBUpdateFile struct {
	WebsiteID uint   `json:"websiteID" validate:"required"`
	Name      string `json:"name" validate:"required"`
	Content   string `json:"content" validate:"required"`
}

type WebsiteRealIP struct {
	WebsiteID uint   `json:"websiteID" validate:"required"`
	Open      bool   `json:"open"`
	IPFrom    string `json:"ipFrom"`
	IPHeader  string `json:"ipHeader"`
	IPOther   string `json:"ipOther"`
}

type ChangeDatabase struct {
	WebsiteID    uint   `json:"websiteID" validate:"required"`
	DatabaseID   uint   `json:"databaseID" `
	DatabaseType string `json:"databaseType" `
}

type WebsiteProxyDel struct {
	ID   uint   `json:"id" validate:"required"`
	Name string `json:"name" validate:"required"`
}

type CrossSiteAccessOp struct {
	WebsiteID uint   `json:"websiteID" validate:"required"`
	Operation string `json:"operation" validate:"required,oneof=Enable Disable"`
}
