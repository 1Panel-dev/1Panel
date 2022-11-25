package model

import "time"

type WebSite struct {
	BaseModel
	Protocol       string          `gorm:"type:varchar(64);not null" json:"protocol"`
	PrimaryDomain  string          `gorm:"type:varchar(128);not null" json:"primaryDomain"`
	Type           string          `gorm:"type:varchar(64);not null" json:"type"`
	Alias          string          `gorm:"type:varchar(128);not null" json:"alias"`
	Remark         string          `gorm:"type:longtext;" json:"remark"`
	Status         string          `gorm:"type:varchar(64);not null" json:"status"`
	ExpireDate     time.Time       `json:"expireDate"`
	AppInstallID   uint            `gorm:"type:integer" json:"appInstallId"`
	WebSiteGroupID uint            `gorm:"type:integer" json:"webSiteGroupId"`
	WebSiteSSLID   uint            `gorm:"type:integer" json:"webSiteSSLId"`
	Domains        []WebSiteDomain `json:"domains"`
	WebSiteSSL     WebSiteSSL      `json:"webSiteSSL"`
}

func (w WebSite) TableName() string {
	return "websites"
}
