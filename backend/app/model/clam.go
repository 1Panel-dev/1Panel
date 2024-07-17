package model

type Clam struct {
	BaseModel

	Name             string `gorm:"type:varchar(64);not null" json:"name"`
	Status           string `gorm:"type:varchar(64)" json:"status"`
	Path             string `gorm:"type:varchar(64);not null" json:"path"`
	InfectedStrategy string `gorm:"type:varchar(64)" json:"infectedStrategy"`
	InfectedDir      string `gorm:"type:varchar(64)" json:"infectedDir"`
	Spec             string `gorm:"type:varchar(64)" json:"spec"`
	EntryID          int    `gorm:"type:varchar(64)" json:"entryID"`
	Description      string `gorm:"type:varchar(64)" json:"description"`
}
