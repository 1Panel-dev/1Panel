package model

import (
	"time"
)

type Cronjob struct {
	BaseModel

	Name string `gorm:"type:varchar(64);not null" json:"name"`
	Type string `gorm:"type:varchar(64);not null" json:"type"`
	Spec string `gorm:"type:varchar(64);not null" json:"spec"`

	Command        string `gorm:"type:varchar(64)" json:"command"`
	ContainerName  string `gorm:"type:varchar(64)" json:"containerName"`
	Script         string `gorm:"longtext" json:"script"`
	Website        string `gorm:"type:varchar(64)" json:"website"`
	AppID          string `gorm:"type:varchar(64)" json:"appID"`
	DBType         string `gorm:"type:varchar(64)" json:"dbType"`
	DBName         string `gorm:"type:varchar(64)" json:"dbName"`
	URL            string `gorm:"type:varchar(256)" json:"url"`
	SourceDir      string `gorm:"type:varchar(256)" json:"sourceDir"`
	ExclusionRules string `gorm:"longtext" json:"exclusionRules"`

	// 已废弃
	KeepLocal   bool   `gorm:"type:varchar(64)" json:"keepLocal"`
	TargetDirID uint64 `gorm:"type:decimal" json:"targetDirID"`

	BackupAccounts  string `gorm:"type:varchar(64)" json:"backupAccounts"`
	DefaultDownload string `gorm:"type:varchar(64)" json:"defaultDownload"`
	RetainCopies    uint64 `gorm:"type:decimal" json:"retainCopies"`

	Status   string       `gorm:"type:varchar(64)" json:"status"`
	EntryIDs string       `gorm:"type:varchar(64)" json:"entryIDs"`
	Records  []JobRecords `json:"records"`
	Secret   string       `gorm:"type:varchar(64)" json:"secret"`
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
