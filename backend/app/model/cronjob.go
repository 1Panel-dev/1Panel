package model

import "time"

type Cronjob struct {
	BaseModel

	Name     string `gorm:"type:varchar(64);not null;unique" json:"name"`
	Type     string `gorm:"type:varchar(64);not null" json:"type"`
	SpecType string `gorm:"type:varchar(64);not null" json:"specType"`
	Spec     string `gorm:"type:varchar(64);not null" json:"spec"`
	Week     uint64 `gorm:"type:decimal" json:"week"`
	Day      uint64 `gorm:"type:decimal" json:"day"`
	Hour     uint64 `gorm:"type:decimal" json:"hour"`
	Minute   uint64 `gorm:"type:decimal" json:"minute"`

	Script         string `gorm:"longtext" json:"script"`
	Website        string `gorm:"type:varchar(64)" json:"website"`
	Database       string `gorm:"type:varchar(64)" json:"database"`
	URL            string `gorm:"type:varchar(256)" json:"url"`
	SourceDir      string `gorm:"type:varchar(256)" json:"sourceDir"`
	TargetDirID    uint64 `gorm:"type:decimal" json:"targetDirID"`
	ExclusionRules string `gorm:"longtext" json:"exclusionRules"`
	RetainDays     uint64 `gorm:"type:decimal" json:"retainDays"`

	Status  string       `gorm:"type:varchar(64)" json:"status"`
	EntryID uint64       `gorm:"type:decimal" json:"entryID"`
	Records []JobRecords `json:"records"`
}

type JobRecords struct {
	BaseModel

	CronjobID uint      `gorm:"type:varchar(64);not null" json:"cronjobID"`
	StartTime time.Time `gorm:"type:datetime" json:"startTime"`
	Interval  float64   `gorm:"type:float" json:"interval"`
	Records   string    `gorm:"longtext" json:"records"`
	Status    string    `gorm:"type:varchar(64)" json:"status"`
	Message   string    `gorm:"longtext" json:"message"`
}
