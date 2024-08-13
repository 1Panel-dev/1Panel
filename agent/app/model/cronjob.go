package model

import (
	"time"
)

type Cronjob struct {
	BaseModel

	Name string `gorm:"not null" json:"name"`
	Type string `gorm:"not null" json:"type"`
	Spec string `gorm:"not null" json:"spec"`

	Command        string `json:"command"`
	ContainerName  string `json:"containerName"`
	Script         string `json:"script"`
	Website        string `json:"website"`
	AppID          string `json:"appID"`
	DBType         string `json:"dbType"`
	DBName         string `json:"dbName"`
	URL            string `json:"url"`
	SourceDir      string `json:"sourceDir"`
	ExclusionRules string `json:"exclusionRules"`

	SourceAccountIDs  string `json:"sourceAccountsIDs"`
	DownloadAccountID uint   `json:"downloadAccountID"`
	RetainCopies      uint64 `json:"retainCopies"`

	Status   string       `json:"status"`
	EntryIDs string       `json:"entryIDs"`
	Records  []JobRecords `json:"records"`
	Secret   string       `json:"secret"`
}

type JobRecords struct {
	BaseModel

	CronjobID uint      `json:"cronjobID"`
	StartTime time.Time `json:"startTime"`
	Interval  float64   `json:"interval"`
	Records   string    `json:"records"`
	FromLocal bool      `json:"source"`
	File      string    `json:"file"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
}
