package model

type ComposeTemplate struct {
	BaseModel

	Name        string `gorm:"type:varchar(64);not null;unique" json:"name"`
	From        string `gorm:"type:varchar(64);not null" json:"from"`
	Description string `gorm:"type:varchar(256)" json:"description"`
	Path        string `gorm:"type:varchar(64)" json:"path"`
	Content     string `gorm:"type:longtext" json:"content"`
}

type Compose struct {
	BaseModel

	Name string `gorm:"type:varchar(256)" json:"name"`
}
