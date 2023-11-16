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
}

type WebsiteDNSReq struct {
	Domains       []string `json:"domains" validate:"required"`
	AcmeAccountID uint     `json:"acmeAccountId"  validate:"required"`
}

type WebsiteSSLRenew struct {
	SSLID uint `json:"SSLId" validate:"required"`
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

type WebsiteSSLUpdate struct {
	ID        uint `json:"id" validate:"required"`
	AutoRenew bool `json:"autoRenew"`
}

type WebsiteSSLUpload struct {
	PrivateKey      string `json:"privateKey"`
	Certificate     string `json:"certificate"`
	PrivateKeyPath  string `json:"privateKeyPath"`
	CertificatePath string `json:"certificatePath"`
	Type            string `json:"type" validate:"required,oneof=paste local"`
}
