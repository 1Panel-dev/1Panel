package dto

import (
	"time"
)

type CommonBackup struct {
	Type       string `json:"type" validate:"required,oneof=app mysql mariadb redis website postgresql"`
	Name       string `json:"name"`
	DetailName string `json:"detailName"`
	Secret     string `json:"secret"`
	TaskID     string `json:"taskID"`
}
type CommonRecover struct {
	DownloadAccountID uint   `json:"downloadAccountID" validate:"required"`
	Type              string `json:"type" validate:"required,oneof=app mysql mariadb redis website postgresql"`
	Name              string `json:"name"`
	DetailName        string `json:"detailName"`
	File              string `json:"file"`
	Secret            string `json:"secret"`
	TaskID            string `json:"taskID"`
	BackupRecordID    uint   `json:"backupRecordID"`
}

type RecordSearch struct {
	PageInfo
	Type       string `json:"type" validate:"required"`
	Name       string `json:"name"`
	DetailName string `json:"detailName"`
}

type RecordSearchByCronjob struct {
	PageInfo
	CronjobID uint `json:"cronjobID" validate:"required"`
}

type BackupRecords struct {
	ID                uint      `json:"id"`
	CreatedAt         time.Time `json:"createdAt"`
	AccountType       string    `json:"accountType"`
	AccountName       string    `json:"accountName"`
	DownloadAccountID uint      `json:"downloadAccountID"`
	FileDir           string    `json:"fileDir"`
	FileName          string    `json:"fileName"`
	Size              int64     `json:"size"`
}

type DownloadRecord struct {
	DownloadAccountID uint   `json:"downloadAccountID" validate:"required"`
	FileDir           string `json:"fileDir" validate:"required"`
	FileName          string `json:"fileName" validate:"required"`
}
