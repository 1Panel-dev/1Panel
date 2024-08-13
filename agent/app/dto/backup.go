package dto

import (
	"time"
)

type BackupInfo struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type" validate:"required"`
	Bucket     string `json:"bucket"`
	AccessKey  string `json:"accessKey"`
	Credential string `json:"credential"`
	BackupPath string `json:"backupPath"`
	Vars       string `json:"vars" validate:"required"`
}

type CommonBackup struct {
	Type       string `json:"type" validate:"required,oneof=app mysql mariadb redis website postgresql"`
	Name       string `json:"name"`
	DetailName string `json:"detailName"`
	Secret     string `json:"secret"`
}
type CommonRecover struct {
	BackupAccountID uint   `json:"backupAccountID" validate:"required"`
	Type            string `json:"type" validate:"required,oneof=app mysql mariadb redis website postgresql"`
	Name            string `json:"name"`
	DetailName      string `json:"detailName"`
	File            string `json:"file"`
	Secret          string `json:"secret"`
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
	ID         uint      `json:"id"`
	CreatedAt  time.Time `json:"createdAt"`
	Source     string    `json:"source"`
	BackupType string    `json:"backupType"`
	FileDir    string    `json:"fileDir"`
	FileName   string    `json:"fileName"`
	Size       int64     `json:"size"`
}

type DownloadRecord struct {
	DownloadAccountID uint   `json:"downloadAccountID" validate:"required"`
	FileDir           string `json:"fileDir" validate:"required"`
	FileName          string `json:"fileName" validate:"required"`
}
