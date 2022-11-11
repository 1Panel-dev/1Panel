package model

import "time"

type WebSiteSSL struct {
	BaseModel
	PrivateKey string    `gorm:"type:longtext;not null" json:"privateKey"`
	Pem        string    `gorm:"type:longtext;not null" json:"pem"`
	Domain     string    `gorm:"type:varchar(256);not null" json:"domain"`
	CertURL    string    `gorm:"type:varchar(256);not null" json:"certURL"`
	Type       string    `gorm:"type:varchar(64);not null" json:"type"`
	IssuerName string    `gorm:"type:varchar(64);not null" json:"issuerName"`
	ExpireDate time.Time `json:"expireDate"`
	StartDate  time.Time `json:"startDate"`
}

func (w WebSiteSSL) TableName() string {
	return "website_ssls"
}
