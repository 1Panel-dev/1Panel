package model

import "time"

type Website struct {
	BaseModel
	Protocol      string    `gorm:"type:varchar;not null" json:"protocol"`
	PrimaryDomain string    `gorm:"type:varchar;not null" json:"primaryDomain"`
	Type          string    `gorm:"type:varchar;not null" json:"type"`
	Alias         string    `gorm:"type:varchar;not null" json:"alias"`
	Remark        string    `gorm:"type:longtext;" json:"remark"`
	Status        string    `gorm:"type:varchar;not null" json:"status"`
	HttpConfig    string    `gorm:"type:varchar;not null" json:"httpConfig"`
	ExpireDate    time.Time `json:"expireDate"`

	Proxy         string `gorm:"type:varchar;" json:"proxy"`
	ProxyType     string `gorm:"type:varchar;" json:"proxyType"`
	SiteDir       string `gorm:"type:varchar;" json:"siteDir"`
	ErrorLog      bool   `json:"errorLog"`
	AccessLog     bool   `json:"accessLog"`
	DefaultServer bool   `json:"defaultServer"`
	IPV6          bool   `json:"IPV6"`
	Rewrite       string `gorm:"type:varchar" json:"rewrite"`

	WebsiteGroupID uint `gorm:"type:integer" json:"webSiteGroupId"`
	WebsiteSSLID   uint `gorm:"type:integer" json:"webSiteSSLId"`
	RuntimeID      uint `gorm:"type:integer" json:"runtimeID"`
	AppInstallID   uint `gorm:"type:integer" json:"appInstallId"`
	FtpID          uint `gorm:"type:integer" json:"ftpId"`

	User  string `gorm:"type:varchar;" json:"user"`
	Group string `gorm:"type:varchar;" json:"group"`

	Domains    []WebsiteDomain `json:"domains" gorm:"-:migration"`
	WebsiteSSL WebsiteSSL      `json:"webSiteSSL" gorm:"-:migration"`
}

func (w Website) TableName() string {
	return "websites"
}
