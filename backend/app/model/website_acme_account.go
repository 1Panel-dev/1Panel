package model

type WebsiteAcmeAccount struct {
	BaseModel
	Email      string `gorm:"type:varchar(256);not null" json:"email"`
	URL        string `gorm:"type:varchar(256);not null" json:"url"`
	PrivateKey string `gorm:"type:longtext;not null" json:"_"`
}

func (w WebsiteAcmeAccount) TableName() string {
	return "website_acme_accounts"
}
