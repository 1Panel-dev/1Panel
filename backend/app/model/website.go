package model

import "time"

type Website struct {
	BaseModel
	Protocol       string          `gorm:"type:varchar(64);not null" json:"protocol"`
	PrimaryDomain  string          `gorm:"type:varchar(128);not null" json:"primaryDomain"`
	Type           string          `gorm:"type:varchar(64);not null" json:"type"`
	Alias          string          `gorm:"type:varchar(128);not null" json:"alias"`
	Remark         string          `gorm:"type:longtext;" json:"remark"`
	Status         string          `gorm:"type:varchar(64);not null" json:"status"`
	HttpConfig     string          `gorm:"type:varchar(64);not null" json:"httpConfig"`
	ExpireDate     time.Time       `json:"expireDate"`
	AppInstallID   uint            `gorm:"type:integer" json:"appInstallId"`
	WebsiteGroupID uint            `gorm:"type:integer" json:"webSiteGroupId"`
	WebsiteSSLID   uint            `gorm:"type:integer" json:"webSiteSSLId"`
	Proxy          string          `gorm:"type:varchar(128);not null" json:"proxy"`
	ErrorLog       bool            `json:"errorLog"`
	AccessLog      bool            `json:"accessLog"`
	DefaultServer  bool            `json:"defaultServer"`
	RuntimeID      uint            `gorm:"type:integer" json:"runtimeID"`
	Domains        []WebsiteDomain `json:"domains" gorm:"-:migration"`
	WebsiteSSL     WebsiteSSL      `json:"webSiteSSL" gorm:"-:migration"`
}

func (w Website) TableName() string {
	return "websites"
}
