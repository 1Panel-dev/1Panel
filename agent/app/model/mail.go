package model

type MailDomain struct {
	BaseModel
	Name       string `gorm:"not null;uniqueIndex" json:"name"`
	Status     string `gorm:"not null;default:'active'" json:"status"`
	DkimKey    string `gorm:"type:text" json:"dkimKey"`
	DnsAutoGen bool   `gorm:"default:true" json:"dnsAutoGen"`
}

type MailAccount struct {
	BaseModel
	DomainID uint   `gorm:"not null;index" json:"domainID"`
	Username string `gorm:"not null" json:"username"`
	Email    string `gorm:"not null;uniqueIndex" json:"email"`
	Quota    int    `gorm:"default:0" json:"quota"`
	Status   string `gorm:"not null;default:'active'" json:"status"`
}

type MailAlias struct {
	BaseModel
	DomainID uint   `gorm:"not null;index" json:"domainID"`
	Source   string `gorm:"not null" json:"source"`
	Target   string `gorm:"not null" json:"target"`
}
