package model

type ComposeTemplate struct {
	BaseModel

	Name        string `gorm:"not null;unique" json:"name"`
	Description string `json:"description"`
	Content     string `json:"content"`
}

type Compose struct {
	BaseModel

	Name string `json:"name"`
}
