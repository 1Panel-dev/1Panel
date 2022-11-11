package model

type WebSiteDomain struct {
	BaseModel
	WebSiteID uint   `gorm:"type:varchar(64);not null" json:"web_site_id"`
	Domain    string `gorm:"type:varchar(256);not null" json:"domain"`
	Port      int    `gorm:"type:integer" json:"port"`
}

func (w WebSiteDomain) TableName() string {
	return "website_domains"
}
