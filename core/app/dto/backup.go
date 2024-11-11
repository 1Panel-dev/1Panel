package dto

import "time"

type BackupOperate struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type" validate:"required"`
	Bucket     string `json:"bucket"`
	AccessKey  string `json:"accessKey"`
	Credential string `json:"credential"`
	BackupPath string `json:"backupPath"`
	Vars       string `json:"vars" validate:"required"`

	RememberAuth bool `json:"rememberAuth"`
}

type BackupOption struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type BackupInfo struct {
	ID         uint      `json:"id"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	Bucket     string    `json:"bucket"`
	AccessKey  string    `json:"accessKey"`
	Credential string    `json:"credential"`
	BackupPath string    `json:"backupPath"`
	Vars       string    `json:"vars"`
	CreatedAt  time.Time `json:"createdAt"`

	RememberAuth bool `json:"rememberAuth"`
}

type BackupClientInfo struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectUri  string `json:"redirect_uri"`
}

type ForBuckets struct {
	Type       string `json:"type" validate:"required"`
	AccessKey  string `json:"accessKey"`
	Credential string `json:"credential" validate:"required"`
	Vars       string `json:"vars" validate:"required"`
}
