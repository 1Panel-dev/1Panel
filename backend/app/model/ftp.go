package model

type Ftp struct {
	BaseModel

	User        string `gorm:"type:varchar(64);not null" json:"user"`
	Password    string `gorm:"type:varchar(64);not null" json:"password"`
	Status      string `gorm:"type:varchar(64);not null" json:"status"`
	Path        string `gorm:"type:varchar(64);not null" json:"path"`
	Description string `gorm:"type:varchar(64);not null" json:"description"`
}
