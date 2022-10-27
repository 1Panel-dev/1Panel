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

type BackupSearch struct {
	PageInfo
	Type       string `json:"type" validate:"required,oneof=website mysql"`
	Name       string `json:"name" validate:"required"`
	DetailName string `json:"detailName"`
}

type BackupRecords struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	Source    string    `json:"source"`
	FileDir   string    `json:"fileDir"`
	FileName  string    `json:"fileName"`
}

type ForBuckets struct {
	Type       string `json:"type" validate:"required"`
	Credential string `json:"credential" validate:"required"`
	Vars       string `json:"vars" validate:"required"`
}
