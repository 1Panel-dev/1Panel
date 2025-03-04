package model

import (
	"time"
)

type WebsiteSSL struct {
	BaseModel
	PrimaryDomain string    `json:"primaryDomain"`
	PrivateKey    string    `json:"privateKey"`
	Pem           string    `json:"pem"`
	Domains       string    `json:"domains"`
	CertURL       string    `json:"certURL"`
	Type          string    `json:"type"`
	Provider      string    `json:"provider"`
	Organization  string    `json:"organization"`
	DnsAccountID  uint      `json:"dnsAccountId"`
	AcmeAccountID uint      `gorm:"column:acme_account_id" json:"acmeAccountId" `
	CaID          uint      `json:"caId"`
	AutoRenew     bool      `json:"autoRenew"`
	ExpireDate    time.Time `json:"expireDate"`
	StartDate     time.Time `json:"startDate"`
	Status        string    `json:"status"`
	Message       string    `json:"message"`
	KeyType       string    `json:"keyType"`
	PushDir       bool      `json:"pushDir"`
	Dir           string    `json:"dir"`
	Description   string    `json:"description"`
	SkipDNS       bool      `json:"skipDNS"`
	Nameserver1   string    `json:"nameserver1"`
	Nameserver2   string    `json:"nameserver2"`
	DisableCNAME  bool      `json:"disableCNAME"`
	ExecShell     bool      `json:"execShell"`
	Shell         string    `json:"shell"`
}

func (w WebsiteSSL) TableName() string {
	return "website_ssls"
}

type WebsiteCA struct {
	BaseModel
	CSR        string `gorm:"not null;" json:"csr"`
	Name       string `gorm:"not null;" json:"name"`
	PrivateKey string `gorm:"not null" json:"privateKey"`
	KeyType    string `gorm:"not null;default:2048" json:"keyType"`
}
