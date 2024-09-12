package model

type ComposeTemplate struct {
	BaseModel

	Name        string `gorm:"type:varchar(64);not null;unique" json:"name"`
	Description string `gorm:"type:varchar(256)" json:"description"`
	Content     string `gorm:"type:longtext" json:"content"`
}

type Compose struct {
	BaseModel

	Name string `gorm:"type:varchar(256)" json:"name"`
	Path string `gorm:"type:varchar(256)" json:"path"`
}
