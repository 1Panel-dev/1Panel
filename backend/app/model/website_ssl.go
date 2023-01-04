package model

import "time"

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
	AutoRenew     bool      `gorm:"type:varchar(64);not null" json:"autoRenew"`
	ExpireDate    time.Time `json:"expireDate"`
	StartDate     time.Time `json:"startDate"`

	AcmeAccount WebsiteAcmeAccount `json:"acmeAccount"`
}

func (w WebsiteSSL) TableName() string {
	return "website_ssls"
}
