package response

import (
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/model"
)

type McpServersRes struct {
	Items []McpServerDTO `json:"items"`
	Total int64          `json:"total"`
}

type McpServerDTO struct {
	model.McpServer
	Environments []request.Environment `json:"environments"`
	Volumes      []request.Volume      `json:"volumes"`
}

type McpServerStatusDTO struct {
	ID      uint   `json:"id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type McpBindDomainRes struct {
	Domain        string   `json:"domain"`
	SSLID         uint     `json:"sslID"`
	AcmeAccountID uint     `json:"acmeAccountID"`
	AllowIPs      []string `json:"allowIPs"`
	WebsiteID     uint     `json:"websiteID"`
	ConnUrl       string   `json:"connUrl"`
}

type McpServerConnectionTestRes struct {
	Success         bool   `json:"success"`
	Endpoint        string `json:"endpoint"`
	OutputTransport string `json:"outputTransport"`
	ProtocolVersion string `json:"protocolVersion,omitempty"`
	Message         string `json:"message"`
}
