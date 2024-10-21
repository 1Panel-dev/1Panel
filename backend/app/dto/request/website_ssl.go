package request

import "github.com/1Panel-dev/1Panel/backend/app/dto"

type WebsiteSSLSearch struct {
	dto.PageInfo
	AcmeAccountID string `json:"acmeAccountID"`
}

type WebsiteSSLCreate struct {
	PrimaryDomain string `json:"primaryDomain" validate:"required"`
	OtherDomains  string `json:"otherDomains"`
	Provider      string `json:"provider" validate:"required"`
	AcmeAccountID uint   `json:"acmeAccountId" validate:"required"`
	DnsAccountID  uint   `json:"dnsAccountId"`
	AutoRenew     bool   `json:"autoRenew"`
	KeyType       string `json:"keyType"`
	Apply         bool   `json:"apply"`
	PushDir       bool   `json:"pushDir"`
	Dir           string `json:"dir"`
	ID            uint   `json:"id"`
	Description   string `json:"description"`
	DisableCNAME  bool   `json:"disableCNAME"`
	SkipDNS       bool   `json:"skipDNS"`
	Nameserver1   string `json:"nameserver1"`
	Nameserver2   string `json:"nameserver2"`
	ExecShell     bool   `json:"execShell"`
	Shell         string `json:"shell"`
}

type WebsiteDNSReq struct {
	Domains       []string `json:"domains" validate:"required"`
	AcmeAccountID uint     `json:"acmeAccountId"  validate:"required"`
}

type WebsiteSSLRenew struct {
	SSLID uint `json:"SSLId" validate:"required"`
}

type WebsiteSSLApply struct {
	ID           uint     `json:"ID" validate:"required"`
	SkipDNSCheck bool     `json:"skipDNSCheck"`
	Nameservers  []string `json:"nameservers"`
	DisableLog   bool     `json:"disableLog"`
}

type WebsiteAcmeAccountCreate struct {
	Email      string `json:"email" validate:"required"`
	Type       string `json:"type" validate:"required,oneof=letsencrypt zerossl buypass google"`
	KeyType    string `json:"keyType" validate:"required,oneof=P256 P384 2048 3072 4096 8192"`
	EabKid     string `json:"eabKid"`
	EabHmacKey string `json:"eabHmacKey"`
}

type WebsiteDnsAccountCreate struct {
	Name          string            `json:"name" validate:"required"`
	Type          string            `json:"type" validate:"required"`
	Authorization map[string]string `json:"authorization" validate:"required"`
}

type WebsiteDnsAccountUpdate struct {
	ID            uint              `json:"id" validate:"required"`
	Name          string            `json:"name" validate:"required"`
	Type          string            `json:"type" validate:"required"`
	Authorization map[string]string `json:"authorization" validate:"required"`
}

type WebsiteResourceReq struct {
	ID uint `json:"id" validate:"required"`
}

type WebsiteBatchDelReq struct {
	IDs []uint `json:"ids" validate:"required"`
}

type WebsiteSSLUpdate struct {
	ID            uint   `json:"id" validate:"required"`
	AutoRenew     bool   `json:"autoRenew"`
	Description   string `json:"description"`
	PrimaryDomain string `json:"primaryDomain" validate:"required"`
	OtherDomains  string `json:"otherDomains"`
	Provider      string `json:"provider" validate:"required"`
	AcmeAccountID uint   `json:"acmeAccountId"`
	DnsAccountID  uint   `json:"dnsAccountId"`
	KeyType       string `json:"keyType"`
	Apply         bool   `json:"apply"`
	PushDir       bool   `json:"pushDir"`
	Dir           string `json:"dir"`
	DisableCNAME  bool   `json:"disableCNAME"`
	SkipDNS       bool   `json:"skipDNS"`
	Nameserver1   string `json:"nameserver1"`
	Nameserver2   string `json:"nameserver2"`
	ExecShell     bool   `json:"execShell"`
	Shell         string `json:"shell"`
}

type WebsiteSSLUpload struct {
	PrivateKey      string `json:"privateKey"`
	Certificate     string `json:"certificate"`
	PrivateKeyPath  string `json:"privateKeyPath"`
	CertificatePath string `json:"certificatePath"`
	Type            string `json:"type" validate:"required,oneof=paste local"`
	SSLID           uint   `json:"sslID"`
	Description     string `json:"description"`
}

type WebsiteCASearch struct {
	dto.PageInfo
}

type WebsiteCACreate struct {
	CommonName       string `json:"commonName" validate:"required"`
	Country          string `json:"country" validate:"required"`
	Organization     string `json:"organization" validate:"required"`
	OrganizationUint string `json:"organizationUint"`
	Name             string `json:"name" validate:"required"`
	KeyType          string `json:"keyType" validate:"required,oneof=P256 P384 2048 3072 4096 8192"`
	Province         string `json:"province" `
	City             string `json:"city"`
}

type WebsiteCAObtain struct {
	ID          uint   `json:"id" validate:"required"`
	Domains     string `json:"domains" validate:"required"`
	KeyType     string `json:"keyType" validate:"required,oneof=P256 P384 2048 3072 4096 8192"`
	Time        int    `json:"time" validate:"required"`
	Unit        string `json:"unit" validate:"required"`
	PushDir     bool   `json:"pushDir"`
	Dir         string `json:"dir"`
	AutoRenew   bool   `json:"autoRenew"`
	Renew       bool   `json:"renew"`
	SSLID       uint   `json:"sslID"`
	Description string `json:"description"`
	ExecShell   bool   `json:"execShell"`
	Shell       string `json:"shell"`
}

type WebsiteCARenew struct {
	SSLID uint `json:"SSLID" validate:"required"`
}
