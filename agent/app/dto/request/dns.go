package request

import "github.com/1Panel-dev/1Panel/agent/app/dto"

type DnsZoneSearch struct {
	dto.PageInfo
	Info    string `json:"info"`
	OrderBy string `json:"orderBy" validate:"oneof=name created_at createdAt status"`
	Order   string `json:"order" validate:"oneof=null ascending descending"`
}

type DnsZoneCreate struct {
	Name       string `json:"name" validate:"required"`
	Type       string `json:"type" validate:"required,oneof=Native Master Slave"`
	MasterIP   string `json:"masterIP"`
	SoaPrimary string `json:"soaPrimary"`
	SoaEmail   string `json:"soaEmail" validate:"required"`
	DefaultTTL int    `json:"defaultTTL"`
}

type DnsZoneUpdate struct {
	ID         uint   `json:"id" validate:"required"`
	SoaPrimary string `json:"soaPrimary"`
	SoaEmail   string `json:"soaEmail"`
	SoaRefresh int    `json:"soaRefresh"`
	SoaRetry   int    `json:"soaRetry"`
	SoaExpire  int    `json:"soaExpire"`
	SoaTTL     int    `json:"soaTTL"`
	DefaultTTL int    `json:"defaultTTL"`
}

type DnsRecordSearch struct {
	dto.PageInfo
	ZoneID uint   `json:"zoneID" validate:"required"`
	Type   string `json:"type"`
	Info   string `json:"info"`
}

type DnsRecordCreate struct {
	ZoneID   uint   `json:"zoneID" validate:"required"`
	Name     string `json:"name" validate:"required"`
	Type     string `json:"type" validate:"required,oneof=A AAAA CNAME MX TXT NS SRV CAA"`
	Content  string `json:"content" validate:"required"`
	TTL      int    `json:"ttl"`
	Priority int    `json:"priority"`
}

type DnsRecordUpdate struct {
	ID       uint   `json:"id" validate:"required"`
	Name     string `json:"name" validate:"required"`
	Type     string `json:"type" validate:"required,oneof=A AAAA CNAME MX TXT NS SRV CAA"`
	Content  string `json:"content" validate:"required"`
	TTL      int    `json:"ttl"`
	Priority int    `json:"priority"`
	Disabled bool   `json:"disabled"`
}

type DnsSetup struct {
	Enable bool `json:"enable"`
}

type DnsSync struct {
	ID uint `json:"id" validate:"required"`
}
