package model

import "time"

type Task struct {
	ID             string    `gorm:"primarykey;" json:"id"`
	Name           string    `json:"name"`
	Type           string    `json:"type"`
	Operate        string    `json:"operate"`
	LogFile        string    `json:"logFile"`
	Status         string    `json:"status"`
	ErrorMsg       string    `json:"errorMsg"`
	OperationLogID uint      `json:"operationLogID"`
	ResourceID     uint      `json:"resourceID"`
	CurrentStep    string    `json:"currentStep"`
	EndAt          time.Time `json:"endAt"`
	CreatedAt      time.Time `json:"createdAt"`
}
