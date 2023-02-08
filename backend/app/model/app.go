package model

type App struct {
	BaseModel
	Name               string      `json:"name" gorm:"type:varchar(64);not null"`
	Key                string      `json:"key" gorm:"type:varchar(64);not null;uniqueIndex"`
	ShortDesc          string      `json:"shortDesc" gorm:"type:longtext;"`
	Icon               string      `json:"icon" gorm:"type:longtext;"`
	Type               string      `json:"type" gorm:"type:varchar(64);not null"`
	Status             string      `json:"status" gorm:"type:varchar(64);not null"`
	Required           string      `json:"required" gorm:"type:varchar(64);not null"`
	CrossVersionUpdate bool        `json:"crossVersionUpdate"`
	Limit              int         `json:"limit" gorm:"type:Integer;not null"`
	Website            string      `json:"website" gorm:"type:varchar(64);not null"`
	Github             string      `json:"github" gorm:"type:varchar(64);not null"`
	Document           string      `json:"document" gorm:"type:varchar(64);not null"`
	Recommend          int         `json:"recommend" gorm:"type:Integer;not null"`
	Details            []AppDetail `json:"-"`
	TagsKey            []string    `json:"-" gorm:"-"`
	AppTags            []AppTag    `json:"-" `
}
