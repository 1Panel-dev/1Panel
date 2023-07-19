package model

type Database struct {
	BaseModel
	Name        string `json:"name" gorm:"type:varchar(64);not null"`
	Type        string `json:"type" gorm:"type:varchar(64);not null"`
	Address     string `json:"address" gorm:"type:varchar(64);not null"`
	Port        uint   `json:"port" gorm:"type:decimal;not null"`
	Username    string `json:"username" gorm:"type:varchar(64)"`
	Password    string `json:"password" gorm:"type:varchar(64)"`
	Format      string `json:"format" gorm:"type:varchar(64)"`
	Description string `json:"description" gorm:"type:varchar(256);"`
}
