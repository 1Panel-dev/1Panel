package request

type DnsZoneSearch struct {
	AccountID uint `json:"accountId" validate:"required"`
}

type DnsRecordSearch struct {
	AccountID uint   `json:"accountId" validate:"required"`
	ZoneID    string `json:"zoneId" validate:"required"`
	ZoneName  string `json:"zoneName" validate:"required"`
}

type DnsRecordOperate struct {
	AccountID uint   `json:"accountId" validate:"required"`
	ZoneID    string `json:"zoneId" validate:"required"`
	ZoneName  string `json:"zoneName" validate:"required"`
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type" validate:"required"`
	Value     string `json:"value" validate:"required"`
	TTL       int    `json:"ttl"`
	Priority  int    `json:"priority"`
	Proxied   bool   `json:"proxied"`
	Remark    string `json:"remark"`
}

type DnsRecordDelete struct {
	AccountID uint   `json:"accountId" validate:"required"`
	ZoneID    string `json:"zoneId" validate:"required"`
	ZoneName  string `json:"zoneName" validate:"required"`
	ID        string `json:"id" validate:"required"`
	Name      string `json:"name"`
	Type      string `json:"type"`
}
