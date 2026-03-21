package response

import "github.com/1Panel-dev/1Panel/agent/app/model"

type DnsZoneRes struct {
	model.DnsZone
	RecordCount int `json:"recordCount"`
}

type DnsRecordRes struct {
	model.DnsRecord
	ZoneName string `json:"zoneName"`
}

type DnsStatus struct {
	Enabled       bool   `json:"enabled"`
	ContainerUp   bool   `json:"containerUp"`
	ContainerName string `json:"containerName"`
	Version       string `json:"version"`
}
