package model

import (
	"time"
)

// BackupAccounts ---> SourceAccountIDs
// BackupAccounts ---> DownloadAccountID
type Cronjob struct {
	BaseModel

	Name       string `gorm:"not null" json:"name"`
	Type       string `gorm:"not null" json:"type"`
	GroupID    uint   `json:"groupID"`
	SpecCustom bool   `json:"specCustom"`
	Spec       string `gorm:"not null" json:"spec"`

	Executor      string `json:"executor"`
	Command       string `json:"command"`
	ContainerName string `json:"containerName"`
	ScriptMode    string `json:"scriptMode"`
	Script        string `json:"script"`
	User          string `json:"user"`

	ScriptID       uint   `json:"scriptID"`
	Website        string `json:"website"`
	AppID          string `json:"appID"`
	DBType         string `json:"dbType"`
	DBName         string `json:"dbName"`
	URL            string `json:"url"`
	IsDir          bool   `json:"isDir"`
	SourceDir      string `json:"sourceDir"`
	SnapshotRule   string `json:"snapshotRule"`
	ExclusionRules string `json:"exclusionRules"`

	SourceAccountIDs  string `json:"sourceAccountIDs"`
	DownloadAccountID uint   `json:"downloadAccountID"`
	RetryTimes        uint   `json:"retryTimes"`
	Timeout           uint   `json:"timeout"`
	IgnoreErr         bool   `json:"ignoreErr"`
	RetainCopies      uint64 `json:"retainCopies"`

	Status   string       `json:"status"`
	EntryIDs string       `json:"entryIDs"`
	Records  []JobRecords `json:"records"`
	Secret   string       `json:"secret"`
}

type JobRecords struct {
	BaseModel

	CronjobID uint      `json:"cronjobID"`
	TaskID    string    `json:"taskID"`
	StartTime time.Time `json:"startTime"`
	Interval  float64   `json:"interval"`
	Records   string    `json:"records"`
	FromLocal bool      `json:"source"`
	File      string    `json:"file"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
}

type ScriptLibrary struct {
	BaseModel
	Name   string `json:"name" gorm:"not null;"`
	Script string `json:"script" gorm:"not null;"`
}
