package dto

import "time"

type MonitorSearch struct {
	Param     string    `json:"param" validate:"required,oneof=all cpu memory load disk inode io iops network"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	Unit      string    `josn:"unit"`
}
