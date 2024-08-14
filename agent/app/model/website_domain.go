package model

type WebsiteDomain struct {
	BaseModel
	WebsiteID uint   `gorm:"column:website_id;not null;" json:"websiteId"`
	Domain    string `gorm:"not null" json:"domain"`
	SSL       bool   `json:"ssl"`
	Port      int    `json:"port"`
}

func (w WebsiteDomain) TableName() string {
	return "website_domains"
}
