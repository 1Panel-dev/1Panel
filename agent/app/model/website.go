package model

import "time"

type Website struct {
	BaseModel
	Protocol      string    `gorm:"not null" json:"protocol"`
	PrimaryDomain string    `gorm:"not null" json:"primaryDomain"`
	Type          string    `gorm:"not null" json:"type"`
	Alias         string    `gorm:"not null" json:"alias"`
	Remark        string    `json:"remark"`
	Status        string    `gorm:"not null" json:"status"`
	HttpConfig    string    `gorm:"not null" json:"httpConfig"`
	ExpireDate    time.Time `json:"expireDate"`

	Proxy         string `json:"proxy"`
	ProxyType     string `json:"proxyType"`
	SiteDir       string `json:"siteDir"`
	ErrorLog      bool   `json:"errorLog"`
	AccessLog     bool   `json:"accessLog"`
	DefaultServer bool   `json:"defaultServer"`
	IPV6          bool   `json:"IPV6"`
	Rewrite       string `json:"rewrite"`

	WebsiteGroupID  uint `json:"webSiteGroupId"`
	WebsiteSSLID    uint `json:"webSiteSSLId"`
	RuntimeID       uint `json:"runtimeID"`
	AppInstallID    uint `json:"appInstallId"`
	FtpID           uint `json:"ftpId"`
	ParentWebsiteID uint `json:"parentWebsiteID"`

	User  string `json:"user"`
	Group string `json:"group"`

	DbType string `json:"dbType"`
	DbID   uint   `json:"dbID"`

	Favorite bool `json:"favorite"`

	Domains    []WebsiteDomain `json:"domains" gorm:"-:migration"`
	WebsiteSSL WebsiteSSL      `json:"webSiteSSL" gorm:"-:migration"`
}

func (w Website) TableName() string {
	return "websites"
}
