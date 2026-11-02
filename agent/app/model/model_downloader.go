package model

import "time"

type ModelDownloadTask struct {
	DownloadType string `json:"downloadType"`
	BaseModel
	Source         string     `json:"source" gorm:"type:varchar(32);not null;index"`
	RepoID         string     `json:"repoID" gorm:"type:varchar(255);not null;index"`
	Revision       string     `json:"revision" gorm:"type:varchar(128)"`
	ModelName      string     `json:"modelName" gorm:"type:varchar(255);not null;index"`
	TargetDir      string     `json:"targetDir" gorm:"type:text"`
	Status         string     `json:"status" gorm:"type:varchar(32);not null;index"`
	Progress       float64    `json:"progress"`
	TotalSize      int64      `json:"totalSize"`
	DownloadedSize int64      `json:"downloadedSize"`
	ErrorMessage   string     `json:"errorMessage" gorm:"type:text"`
	StartedAt      *time.Time `json:"startedAt"`
	CompletedAt    *time.Time `json:"completedAt"`
}
