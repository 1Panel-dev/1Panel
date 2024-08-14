package model

type Clam struct {
	BaseModel

	Name             string `gorm:"not null" json:"name"`
	Status           string `json:"status"`
	Path             string `gorm:"not null" json:"path"`
	InfectedStrategy string `json:"infectedStrategy"`
	InfectedDir      string `json:"infectedDir"`
	Spec             string `json:"spec"`
	EntryID          int    `json:"entryID"`
	Description      string `json:"description"`
}
