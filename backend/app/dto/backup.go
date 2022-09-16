package dto

import "time"

type BackupOperate struct {
	Name       string `json:"name" validate:"required"`
	Type       string `json:"type" validate:"required"`
	Bucket     string `json:"bucket"`
	Credential string `json:"credential"`
	Vars       string `json:"vars" validate:"required"`
}

type BackupInfo struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Bucket    string    `json:"bucket"`
	Vars      string    `json:"vars"`
}

type ForBuckets struct {
	Type       string `json:"type" validate:"required"`
	Credential string `json:"credential" validate:"required"`
	Vars       string `json:"vars" validate:"required"`
}
