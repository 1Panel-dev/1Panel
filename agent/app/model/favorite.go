package model

type Favorite struct {
	BaseModel
	Name  string `gorm:"not null;" json:"name" `
	Path  string `gorm:"not null;unique" json:"path"`
	Type  string `json:"type"`
	IsDir bool   `json:"isDir"`
	IsTxt bool   `json:"isTxt"`
}
