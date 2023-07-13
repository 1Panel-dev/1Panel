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
	Second   uint64 `gorm:"type:decimal" json:"second"`

	ContainerName  string `gorm:"type:varchar(64)" json:"containerName"`
	Script         string `gorm:"longtext" json:"script"`
	Website        string `gorm:"type:varchar(64)" json:"website"`
	DBName         string `gorm:"type:varchar(64)" json:"dbName"`
	URL            string `gorm:"type:varchar(256)" json:"url"`
	SourceDir      string `gorm:"type:varchar(256)" json:"sourceDir"`
	ExclusionRules string `gorm:"longtext" json:"exclusionRules"`

	KeepLocal    bool   `gorm:"type:varchar(64)" json:"keepLocal"`
	TargetDirID  uint64 `gorm:"type:decimal" json:"targetDirID"`
	RetainCopies uint64 `gorm:"type:decimal" json:"retainCopies"`

	Status  string       `gorm:"type:varchar(64)" json:"status"`
	EntryID uint64       `gorm:"type:decimal" json:"entryID"`
	Records []JobRecords `json:"records"`
}

type JobRecords struct {
	BaseModel

	CronjobID uint      `gorm:"type:decimal" json:"cronjobID"`
	StartTime time.Time `gorm:"type:datetime" json:"startTime"`
	Interval  float64   `gorm:"type:float" json:"interval"`
	Records   string    `gorm:"longtext" json:"records"`
	FromLocal bool      `gorm:"type:varchar(64)" json:"source"`
	File      string    `gorm:"type:varchar(256)" json:"file"`
	Status    string    `gorm:"type:varchar(64)" json:"status"`
	Message   string    `gorm:"longtext" json:"message"`
}
