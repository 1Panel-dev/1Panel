package dto

import (
	"time"
)

type SyncFromMaster struct {
	Name      string `json:"name" validate:"required"`
	Operation string `json:"operation" validate:"required,oneof=create delete update"`
	Data      string `json:"data"`
}

type BackupOption struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type CommonBackup struct {
	Type       string `json:"type" validate:"required,oneof=app mysql mariadb redis website postgresql"`
	Name       string `json:"name"`
	DetailName string `json:"detailName"`
	Secret     string `json:"secret"`
	TaskID     string `json:"taskID"`
	FileName   string `json:"fileName"`
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
