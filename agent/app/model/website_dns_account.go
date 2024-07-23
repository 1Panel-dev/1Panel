package model

type WebsiteDnsAccount struct {
	BaseModel
	Name          string `gorm:"type:varchar(64);not null" json:"name"`
	Type          string `gorm:"type:varchar(64);not null" json:"type"`
	Authorization string `gorm:"type:varchar(256);not null" json:"-"`
}

func (w WebsiteDnsAccount) TableName() string {
	return "website_dns_accounts"
}
