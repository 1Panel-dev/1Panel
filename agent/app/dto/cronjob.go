package dto

import (
	"time"
)

type PageCronjob struct {
	PageInfo
	Info    string `json:"info"`
	OrderBy string `json:"orderBy" validate:"required,oneof=name status createdAt"`
	Order   string `json:"order" validate:"required,oneof=null ascending descending"`
}

type CronjobSpec struct {
	Spec string `json:"spec" validate:"required"`
}

type CronjobCreate struct {
	Name       string `json:"name" validate:"required"`
	Type       string `json:"type" validate:"required"`
	SpecCustom bool   `json:"specCustom"`
	Spec       string `json:"spec" validate:"required"`

	Executor      string `json:"executor"`
	ScriptMode    string `json:"scriptMode"`
	Script        string `json:"script"`
	Command       string `json:"command"`
	ContainerName string `json:"containerName"`
	User          string `json:"user"`

	AppID          string `json:"appID"`
	Website        string `json:"website"`
	ExclusionRules string `json:"exclusionRules"`
	DBType         string `json:"dbType"`
	DBName         string `json:"dbName"`
	URL            string `json:"url"`
	IsDir          bool   `json:"isDir"`
	SourceDir      string `json:"sourceDir"`

	SourceAccountIDs  string `json:"sourceAccountIDs"`
	DownloadAccountID uint   `json:"downloadAccountID"`
	RetainCopies      int    `json:"retainCopies" validate:"number,min=1"`
	Secret            string `json:"secret"`

	AlertCount uint   `json:"alertCount"`
	AlertTitle string `json:"alertTitle"`
}

type CronjobUpdate struct {
	ID         uint   `json:"id" validate:"required"`
	Type       string `json:"type" validate:"required"`
	Name       string `json:"name" validate:"required"`
	SpecCustom bool   `json:"specCustom"`
	Spec       string `json:"spec" validate:"required"`

	Executor      string `json:"executor"`
	ScriptMode    string `json:"scriptMode"`
	Script        string `json:"script"`
	Command       string `json:"command"`
	ContainerName string `json:"containerName"`
	User          string `json:"user"`

	AppID          string `json:"appID"`
	Website        string `json:"website"`
	ExclusionRules string `json:"exclusionRules"`
	DBType         string `json:"dbType"`
	DBName         string `json:"dbName"`
	URL            string `json:"url"`
	IsDir          bool   `json:"isDir"`
	SourceDir      string `json:"sourceDir"`

	SourceAccountIDs  string `json:"sourceAccountIDs"`
	DownloadAccountID uint   `json:"downloadAccountID"`
	RetainCopies      int    `json:"retainCopies" validate:"number,min=1"`
	Secret            string `json:"secret"`

	AlertCount uint   `json:"alertCount"`
	AlertTitle string `json:"alertTitle"`
}

type CronjobUpdateStatus struct {
	ID     uint   `json:"id" validate:"required"`
	Status string `json:"status" validate:"required"`
}

type CronjobDownload struct {
	RecordID        uint `json:"recordID" validate:"required"`
	BackupAccountID uint `json:"backupAccountID" validate:"required"`
}

type CronjobClean struct {
	IsDelete  bool `json:"isDelete"`
	CleanData bool `json:"cleanData"`
	CronjobID uint `json:"cronjobID" validate:"required"`
}

type CronjobBatchDelete struct {
	CleanData bool   `json:"cleanData"`
	IDs       []uint `json:"ids" validate:"required"`
}

type CronjobInfo struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	SpecCustom bool   `json:"specCustom"`
	Spec       string `json:"spec"`

	Executor      string `json:"executor"`
	ScriptMode    string `json:"scriptMode"`
	Script        string `json:"script"`
	Command       string `json:"command"`
	ContainerName string `json:"containerName"`
	User          string `json:"user"`

	AppID          string `json:"appID"`
	Website        string `json:"website"`
	ExclusionRules string `json:"exclusionRules"`
	DBType         string `json:"dbType"`
	DBName         string `json:"dbName"`
	URL            string `json:"url"`
	IsDir          bool   `json:"isDir"`
	SourceDir      string `json:"sourceDir"`
	RetainCopies   int    `json:"retainCopies"`

	SourceAccounts    []string `json:"sourceAccounts"`
	DownloadAccount   string   `json:"downloadAccount"`
	SourceAccountIDs  string   `json:"sourceAccountIDs"`
	DownloadAccountID uint     `json:"downloadAccountID"`

	LastRecordStatus string `json:"lastRecordStatus"`
	LastRecordTime   string `json:"lastRecordTime"`
	Status           string `json:"status"`
	Secret           string `json:"secret"`

	AlertCount uint `json:"alertCount"`
}

type SearchRecord struct {
	PageInfo
	CronjobID int       `json:"cronjobID"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	Status    string    `json:"status"`
}

type Record struct {
	ID         uint   `json:"id"`
	TaskID     string `json:"taskID"`
	StartTime  string `json:"startTime"`
	Records    string `json:"records"`
	Status     string `json:"status"`
	Message    string `json:"message"`
	TargetPath string `json:"targetPath"`
	Interval   int    `json:"interval"`
	File       string `json:"file"`
}
