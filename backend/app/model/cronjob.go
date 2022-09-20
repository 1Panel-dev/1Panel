package model

type Cronjob struct {
	BaseModel

	Name     string `gorm:"type:varchar(64);not null" json:"name"`
	Type     string `gorm:"type:varchar(64);not null" json:"type"`
	SpecType string `gorm:"type:varchar(64);not null" json:"specType"`
	Spec     string `gorm:"type:varchar(64);not null" json:"spec"`
	Week     int    `gorm:"type:varchar(64)" json:"week"`
	Day      int    `gorm:"type:varchar(64)" json:"day"`
	Hour     int    `gorm:"type:varchar(64)" json:"hour"`
	Minute   int    `gorm:"type:varchar(64)" json:"minute"`

	Script         string `gorm:"longtext" json:"script"`
	WebSite        string `gorm:"type:varchar(64)" json:"webSite"`
	ExclusionRules string `gorm:"longtext" json:"exclusionRules"`
	Database       string `gorm:"type:varchar(64)" json:"database"`
	URL            string `gorm:"type:varchar(256)" json:"url"`
	TargetDir      string `gorm:"type:varchar(64)" json:"targetDir"`
	RetainCopies   string `gorm:"type:varchar(64)" json:"retainCopies"`

	Status string `gorm:"type:varchar(64)" json:"status"`
}
