package dto

import (
	"time"
)

type SearchClamWithPage struct {
	PageInfo
	Info    string `json:"info"`
	OrderBy string `json:"orderBy" validate:"required,oneof=name status created_at"`
	Order   string `json:"order" validate:"required,oneof=null ascending descending"`
}

type ClamBaseInfo struct {
	Version  string `json:"version"`
	IsActive bool   `json:"isActive"`
	IsExist  bool   `json:"isExist"`

	FreshVersion  string `json:"freshVersion"`
	FreshIsActive bool   `json:"freshIsActive"`
	FreshIsExist  bool   `json:"freshIsExist"`
}

type ClamInfo struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`

	Name             string `json:"name"`
	Status           string `json:"status"`
	Path             string `json:"path"`
	InfectedStrategy string `json:"infectedStrategy"`
	InfectedDir      string `json:"infectedDir"`
	LastHandleDate   string `json:"lastHandleDate"`
	Spec             string `json:"spec"`
	Description      string `json:"description"`
	AlertCount       uint   `json:"alertCount"`
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

type ClamFileReq struct {
	Tail string `json:"tail"`
	Name string `json:"name" validate:"required"`
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
	Status           string `json:"status"`
	Path             string `json:"path"`
	InfectedStrategy string `json:"infectedStrategy"`
	InfectedDir      string `json:"infectedDir"`
	Spec             string `json:"spec"`
	Description      string `json:"description"`
	AlertCount       uint   `json:"alertCount"`
	AlertTitle       string `json:"alertTitle"`
}

type ClamUpdate struct {
	ID uint `json:"id"`

	Name             string `json:"name"`
	Path             string `json:"path"`
	InfectedStrategy string `json:"infectedStrategy"`
	InfectedDir      string `json:"infectedDir"`
	Spec             string `json:"spec"`
	Description      string `json:"description"`
	AlertCount       uint   `json:"alertCount"`
	AlertTitle       string `json:"alertTitle"`
}

type ClamUpdateStatus struct {
	ID     uint   `json:"id"`
	Status string `json:"status"`
}

type ClamDelete struct {
	RemoveRecord   bool   `json:"removeRecord"`
	RemoveInfected bool   `json:"removeInfected"`
	Ids            []uint `json:"ids" validate:"required"`
}
