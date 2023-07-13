package dto

import "time"

type CronjobCreate struct {
	Name     string `json:"name" validate:"required"`
	Type     string `json:"type" validate:"required"`
	SpecType string `json:"specType" validate:"required"`
	Week     int    `json:"week" validate:"number,max=6,min=0"`
	Day      int    `json:"day" validate:"number"`
	Hour     int    `json:"hour" validate:"number"`
	Minute   int    `json:"minute" validate:"number"`
	Second   int    `json:"second" validate:"number"`

	Script         string `json:"script"`
	ContainerName  string `json:"containerName"`
	Website        string `json:"website"`
	ExclusionRules string `json:"exclusionRules"`
	DBName         string `json:"dbName"`
	URL            string `json:"url"`
	SourceDir      string `json:"sourceDir"`
	KeepLocal      bool   `json:"keepLocal"`
	TargetDirID    int    `json:"targetDirID"`
	RetainCopies   int    `json:"retainCopies" validate:"number,min=1"`
}

type CronjobUpdate struct {
	ID       uint   `json:"id" validate:"required"`
	Name     string `json:"name" validate:"required"`
	SpecType string `json:"specType" validate:"required"`
	Week     int    `json:"week" validate:"number,max=6,min=0"`
	Day      int    `json:"day" validate:"number"`
	Hour     int    `json:"hour" validate:"number"`
	Minute   int    `json:"minute" validate:"number"`
	Second   int    `json:"second" validate:"number"`

	Script         string `json:"script"`
	ContainerName  string `json:"containerName"`
	Website        string `json:"website"`
	ExclusionRules string `json:"exclusionRules"`
	DBName         string `json:"dbName"`
	URL            string `json:"url"`
	SourceDir      string `json:"sourceDir"`
	KeepLocal      bool   `json:"keepLocal"`
	TargetDirID    int    `json:"targetDirID"`
	RetainCopies   int    `json:"retainCopies" validate:"number,min=1"`
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
	CleanData bool `json:"cleanData"`
	CronjobID uint `json:"cronjobID" validate:"required"`
}

type CronjobBatchDelete struct {
	CleanData bool   `json:"cleanData"`
	IDs       []uint `json:"ids"`
}

type CronjobInfo struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	SpecType string `json:"specType"`
	Week     int    `json:"week"`
	Day      int    `json:"day"`
	Hour     int    `json:"hour"`
	Minute   int    `json:"minute"`
	Second   int    `json:"second"`

	Script         string `json:"script"`
	ContainerName  string `json:"containerName"`
	Website        string `json:"website"`
	ExclusionRules string `json:"exclusionRules"`
	DBName         string `json:"dbName"`
	URL            string `json:"url"`
	SourceDir      string `json:"sourceDir"`
	KeepLocal      bool   `json:"keepLocal"`
	TargetDir      string `json:"targetDir"`
	TargetDirID    int    `json:"targetDirID"`
	RetainCopies   int    `json:"retainCopies"`

	LastRecordTime string `json:"lastRecordTime"`
	Status         string `json:"status"`
}

type SearchRecord struct {
	PageInfo
	CronjobID int       `json:"cronjobID"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	Status    string    `json:"status"`
}

type Record struct {
	ID         uint      `json:"id"`
	StartTime  time.Time `json:"startTime"`
	Records    string    `json:"records"`
	Status     string    `json:"status"`
	Message    string    `json:"message"`
	TargetPath string    `json:"targetPath"`
	Interval   int       `json:"interval"`
	File       string    `json:"file"`
}
