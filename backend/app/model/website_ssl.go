package model

import (
	"fmt"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"path"
	"time"
)

type WebsiteSSL struct {
	BaseModel
	PrimaryDomain string    `gorm:"type:varchar(256);not null" json:"primaryDomain"`
	PrivateKey    string    `gorm:"type:longtext;not null" json:"privateKey"`
	Pem           string    `gorm:"type:longtext;not null" json:"pem"`
	Domains       string    `gorm:"type:varchar(256);not null" json:"domains"`
	CertURL       string    `gorm:"type:varchar(256);not null" json:"certURL"`
	Type          string    `gorm:"type:varchar(64);not null" json:"type"`
	Provider      string    `gorm:"type:varchar(64);not null" json:"provider"`
	Organization  string    `gorm:"type:varchar(64);not null" json:"organization"`
	DnsAccountID  uint      `gorm:"type:integer;not null" json:"dnsAccountId"`
	AcmeAccountID uint      `gorm:"type:integer;not null" json:"acmeAccountId"`
	CaID          uint      `gorm:"type:integer;not null;default:0" json:"caId"`
	AutoRenew     bool      `gorm:"type:varchar(64);not null" json:"autoRenew"`
	ExpireDate    time.Time `json:"expireDate"`
	StartDate     time.Time `json:"startDate"`
	Status        string    `gorm:"not null;default:ready" json:"status"`
	Message       string    `json:"message"`
	KeyType       string    `gorm:"not null;default:2048" json:"keyType"`
	PushDir       bool      `gorm:"not null;default:0" json:"pushDir"`
	Dir           string    `json:"dir"`
	Description   string    `json:"description"`

	AcmeAccount WebsiteAcmeAccount `json:"acmeAccount" gorm:"-:migration"`
	DnsAccount  WebsiteDnsAccount  `json:"dnsAccount" gorm:"-:migration"`
	Websites    []Website          `json:"websites" gorm:"-:migration"`
}

func (w WebsiteSSL) TableName() string {
	return "website_ssls"
}

func (w WebsiteSSL) GetLogPath() string {
	return path.Join(constant.SSLLogDir, fmt.Sprintf("%s-ssl-%d.log", w.PrimaryDomain, w.ID))
}
