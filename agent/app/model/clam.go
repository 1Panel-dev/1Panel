package model

import "time"

type Clam struct {
	BaseModel

	Name             string `gorm:"not null" json:"name"`
	Path             string `gorm:"not null" json:"path"`
	InfectedStrategy string `json:"infectedStrategy"`
	InfectedDir      string `json:"infectedDir"`
	Spec             string `json:"spec"`
	RetryTimes       uint   `json:"retryTimes"`
	Timeout          uint   `json:"timeout"`
	EntryID          int    `json:"entryID"`
	Description      string `json:"description"`

	Status      string `json:"status"`
	IsExecuting bool   `json:"isExecuting"`
}

type ClamRecord struct {
	BaseModel

	ClamID        uint      `json:"clamID"`
	TaskID        string    `json:"taskID"`
	StartTime     time.Time `json:"startTime"`
	ScanTime      string    `json:"scanTime"`
	InfectedFiles string    `json:"infectedFiles"`
	TotalError    string    `json:"totalError"`
	Status        string    `json:"status"`
	Message       string    `json:"message"`
}
