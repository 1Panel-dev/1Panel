package dto

import "time"

type MonitorSearch struct {
	Param     string    `json:"param" validate:"required,oneof=all cpu memory load io network"`
	Info      string    `json:"info"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

type MonitorData struct {
	Param string        `json:"param" validate:"required,oneof=cpu memory load io network"`
	Date  []time.Time   `json:"date"`
	Value []interface{} `json:"value"`
}
