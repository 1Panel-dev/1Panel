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

type BackupFile struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Size int64  `json:"size"`
}

type OneDriveInfo struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectUri  string `json:"redirect_uri"`
}

type BackupSearchFile struct {
	Type string `json:"type" validate:"required"`
}

type CommonBackup struct {
	Type       string `json:"type" validate:"required,oneof=app mysql mariadb redis website postgresql"`
	Name       string `json:"name"`
	DetailName string `json:"detailName"`
	Secret     string `json:"secret"`
	FileName   string `json:"fileName"`
}
type CommonRecover struct {
	Source     string `json:"source" validate:"required,oneof=OSS S3 SFTP MINIO LOCAL COS KODO OneDrive WebDAV"`
	Type       string `json:"type" validate:"required,oneof=app mysql mariadb redis website postgresql"`
	Name       string `json:"name"`
	DetailName string `json:"detailName"`
	File       string `json:"file"`
	Secret     string `json:"secret"`
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
}

type DownloadRecord struct {
	Source   string `json:"source" validate:"required,oneof=OSS S3 SFTP MINIO LOCAL COS KODO OneDrive WebDAV"`
	FileDir  string `json:"fileDir" validate:"required"`
	FileName string `json:"fileName" validate:"required"`
}

type ForBuckets struct {
	Type       string `json:"type" validate:"required"`
	AccessKey  string `json:"accessKey"`
	Credential string `json:"credential" validate:"required"`
	Vars       string `json:"vars" validate:"required"`
}
