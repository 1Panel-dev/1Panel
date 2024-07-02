package model

type Clam struct {
	BaseModel

	Name             string `gorm:"type:varchar(64);not null" json:"name"`
	Path             string `gorm:"type:varchar(64);not null" json:"path"`
	InfectedStrategy string `gorm:"type:varchar(64)" json:"infectedStrategy"`
	InfectedDir      string `gorm:"type:varchar(64)" json:"infectedDir"`
	Description      string `gorm:"type:varchar(64)" json:"description"`
}
