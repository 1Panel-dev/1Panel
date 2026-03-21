package model

type DnsZone struct {
	BaseModel
	Name       string `gorm:"not null;uniqueIndex" json:"name"`
	Type       string `gorm:"not null;default:'Native'" json:"type"`
	MasterIP   string `json:"masterIP"`
	SoaPrimary string `json:"soaPrimary"`
	SoaEmail   string `json:"soaEmail"`
	SoaRefresh int    `gorm:"default:10800" json:"soaRefresh"`
	SoaRetry   int    `gorm:"default:3600" json:"soaRetry"`
	SoaExpire  int    `gorm:"default:604800" json:"soaExpire"`
	SoaTTL     int    `gorm:"default:3600" json:"soaTTL"`
	DefaultTTL int    `gorm:"default:3600" json:"defaultTTL"`
	Status     string `gorm:"not null;default:'active'" json:"status"`
	PdnsID     string `json:"pdnsID"`
}

type DnsRecord struct {
	BaseModel
	ZoneID   uint   `gorm:"not null;index" json:"zoneID"`
	Name     string `gorm:"not null" json:"name"`
	Type     string `gorm:"not null" json:"type"`
	Content  string `gorm:"not null" json:"content"`
	TTL      int    `gorm:"default:3600" json:"ttl"`
	Priority int    `json:"priority"`
	Disabled bool   `gorm:"default:false" json:"disabled"`
	AutoGen  bool   `gorm:"default:false" json:"autoGen"`
}
