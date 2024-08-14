package model

type Ftp struct {
	BaseModel

	User        string `gorm:"not null" json:"user"`
	Password    string `gorm:"not null" json:"password"`
	Status      string `gorm:"not null" json:"status"`
	Path        string `gorm:"not null" json:"path"`
	Description string `gorm:"not null" json:"description"`
}
