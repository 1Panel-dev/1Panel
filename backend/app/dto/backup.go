package dto

import "time"

type BackupOperate struct {
	ID         uint   `json:"id"`
	Type       string `json:"type" validate:"required"`
	Bucket     string `json:"bucket"`
	AccessKey  string `json:"accessKey"`
	Credential string `json:"credential"`
	BackupPath string `json:"backupPath"`
	Vars       string `json:"vars" validate:"required"`
}

type BackupInfo struct {
	ID         uint      `json:"id"`
	CreatedAt  time.Time `json:"createdAt"`
	Type       string    `json:"type"`
	Bucket     string    `json:"bucket"`
	BackupPath string    `json:"backupPath"`
	Vars       string    `json:"vars"`
}

type BackupSearch struct {
	PageInfo
	Type       string `json:"type" validate:"required,oneof=website mysql"`
	Name       string `json:"name" validate:"required"`
	DetailName string `json:"detailName"`
}

type BackupSearchFile struct {
	Type string `json:"type" validate:"required"`
}

type CommonBackup struct {
	Type       string `json:"type" validate:"required,oneof=app mysql redis website"`
	Name       string `json:"name"`
	DetailName string `json:"detailName"`
}
type CommonRecover struct {
	Source     string `json:"source" validate:"required,oneof=OSS S3 SFTP MINIO LOCAL COS KODO OneDrive"`
	Type       string `json:"type" validate:"required,oneof=app mysql redis website"`
	Name       string `json:"name"`
	DetailName string `json:"detailName"`
	File       string `json:"file"`
}

type RecordSearch struct {
	PageInfo
	Type       string `json:"type" validate:"required"`
	Name       string `json:"name" validate:"required"`
	DetailName string `json:"detailName"`
}

type BackupRecords struct {
	ID         uint      `json:"id"`
	CreatedAt  time.Time `json:"createdAt"`
	Source     string    `json:"source"`
	BackupType string    `json:"backupType"`
	FileDir    string    `json:"fileDir"`
	FileName   string    `json:"fileName"`
}

type DownloadRecord struct {
	Source   string `json:"source" validate:"required,oneof=OSS S3 SFTP MINIO LOCAL COS KODO OneDrive"`
	FileDir  string `json:"fileDir" validate:"required"`
	FileName string `json:"fileName" validate:"required"`
}

type ForBuckets struct {
	Type       string `json:"type" validate:"required"`
	AccessKey  string `json:"accessKey"`
	Credential string `json:"credential" validate:"required"`
	Vars       string `json:"vars" validate:"required"`
}
