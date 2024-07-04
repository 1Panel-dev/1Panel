package dto

import (
	"time"
)

type ClamBaseInfo struct {
	Version  string `json:"version"`
	IsActive bool   `json:"isActive"`
	IsExist  bool   `json:"isExist"`
}

type ClamInfo struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`

	Name             string `json:"name"`
	Path             string `json:"path"`
	InfectedStrategy string `json:"infectedStrategy"`
	InfectedDir      string `json:"infectedDir"`
	LastHandleDate   string `json:"lastHandleDate"`
	Description      string `json:"description"`
}

type ClamLogSearch struct {
	PageInfo

	ClamID    uint      `json:"clamID"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

type ClamLogReq struct {
	Tail       string `json:"tail"`
	ClamName   string `json:"clamName"`
	RecordName string `json:"recordName"`
}

type ClamLog struct {
	Name          string `json:"name"`
	ScanDate      string `json:"scanDate"`
	ScanTime      string `json:"scanTime"`
	InfectedFiles string `json:"infectedFiles"`
	TotalError    string `json:"totalError"`
	Status        string `json:"status"`
}

type ClamCreate struct {
	Name             string `json:"name"`
	Path             string `json:"path"`
	InfectedStrategy string `json:"infectedStrategy"`
	InfectedDir      string `json:"infectedDir"`
	Description      string `json:"description"`
}

type ClamUpdate struct {
	ID uint `json:"id"`

	Name             string `json:"name"`
	Path             string `json:"path"`
	InfectedStrategy string `json:"infectedStrategy"`
	InfectedDir      string `json:"infectedDir"`
	Description      string `json:"description"`
}

type ClamDelete struct {
	RemoveRecord   bool   `json:"removeRecord"`
	RemoveInfected bool   `json:"removeInfected"`
	Ids            []uint `json:"ids" validate:"required"`
}
