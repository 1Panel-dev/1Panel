package response

import "github.com/1Panel-dev/1Panel/backend/app/model"

type WebsiteSSLDTO struct {
	model.WebsiteSSL
	LogPath string `json:"logPath"`
}

type WebsiteDNSRes struct {
	Key    string `json:"resolve"`
	Value  string `json:"value"`
	Domain string `json:"domain"`
	Err    string `json:"err"`
}

type WebsiteAcmeAccountDTO struct {
	model.WebsiteAcmeAccount
}

type WebsiteDnsAccountDTO struct {
	model.WebsiteDnsAccount
	Authorization map[string]string `json:"authorization"`
}

type WebsiteCADTO struct {
	model.WebsiteCA
	CommonName       string `json:"commonName" `
	Country          string `json:"country"`
	Organization     string `json:"organization"`
	OrganizationUint string `json:"organizationUint"`
	Province         string `json:"province" `
	City             string `json:"city"`
}
