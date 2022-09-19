package dto

import "time"

type BackupOperate struct {
	Type       string `json:"type" validate:"required"`
	Bucket     string `json:"bucket"`
	Credential string `json:"credential"`
	Vars       string `json:"vars" validate:"required"`
}

type BackupInfo struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	Type      string    `json:"type"`
	Bucket    string    `json:"bucket"`
	Vars      string    `json:"vars"`
}

type ForBuckets struct {
	Type       string `json:"type" validate:"required"`
	Credential string `json:"credential" validate:"required"`
	Vars       string `json:"vars" validate:"required"`
}
