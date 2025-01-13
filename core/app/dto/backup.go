package dto

import "time"

type SyncToAgent struct {
	Name      string `json:"name" validate:"required"`
	Operation string `json:"operation" validate:"required,oneof=create delere update"`
	Data      string `json:"data"`
}

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
