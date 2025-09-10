package dto

import (
	"time"
)

type BackupOperate struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type" validate:"required"`
	IsPublic   bool   `json:"isPublic"`
	Bucket     string `json:"bucket"`
	AccessKey  string `json:"accessKey"`
	Credential string `json:"credential"`
	BackupPath string `json:"backupPath"`
	Vars       string `json:"vars" validate:"required"`

	RememberAuth bool `json:"rememberAuth"`
}

type BackupInfo struct {
	ID         uint      `json:"id"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	IsPublic   bool      `json:"isPublic"`
	Bucket     string    `json:"bucket"`
	AccessKey  string    `json:"accessKey"`
	Credential string    `json:"credential"`
	BackupPath string    `json:"backupPath"`
	Vars       string    `json:"vars"`
	CreatedAt  time.Time `json:"createdAt"`

	RememberAuth bool `json:"rememberAuth"`
}

type ForBuckets struct {
	Type       string `json:"type" validate:"required"`
	AccessKey  string `json:"accessKey"`
	Credential string `json:"credential" validate:"required"`
	Vars       string `json:"vars" validate:"required"`
}

type SyncFromMaster struct {
	Name      string `json:"name" validate:"required"`
	Operation string `json:"operation" validate:"required,oneof=create delete update"`
	Data      string `json:"data"`
}

type BackupOption struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	IsPublic bool   `json:"isPublic"`
}

type UploadForRecover struct {
	FilePath  string `json:"filePath"`
	TargetDir string `json:"targetDir"`
}

type CommonBackup struct {
	Type       string `json:"type" validate:"required,oneof=app mysql mariadb redis website postgresql mysql-cluster postgresql-cluster redis-cluster"`
	Name       string `json:"name"`
	DetailName string `json:"detailName"`
	Secret     string `json:"secret"`
	TaskID     string `json:"taskID"`
	FileName   string `json:"fileName"`

	Description string `json:"description"`
}
type CommonRecover struct {
	DownloadAccountID uint   `json:"downloadAccountID" validate:"required"`
	Type              string `json:"type" validate:"required,oneof=app mysql mariadb redis website postgresql mysql-cluster postgresql-cluster redis-cluster"`
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
	TaskID            string    `json:"taskID"`
	Status            string    `json:"status"`
	Message           string    `json:"message"`
	Description       string    `json:"description"`
}

type DownloadRecord struct {
	DownloadAccountID uint   `json:"downloadAccountID" validate:"required"`
	FileDir           string `json:"fileDir" validate:"required"`
	FileName          string `json:"fileName" validate:"required"`
}

type SearchForSize struct {
	PageInfo
	Type       string `json:"type" validate:"required"`
	Name       string `json:"name"`
	DetailName string `json:"detailName"`
	Info       string `json:"info"`
	CronjobID  uint   `json:"cronjobID"`
	OrderBy    string `json:"orderBy"`
	Order      string `json:"order"`
}
type RecordFileSize struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Size int64  `json:"size"`
}
