package dto

import "time"

type ImageRepoCreate struct {
	Name        string `json:"name" validate:"required"`
	DownloadUrl string `json:"downloadUrl"`
	RepoName    string `json:"repoName"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	Auth        bool   `json:"auth"`
}

type ImageRepoUpdate struct {
	ID          uint   `json:"id"`
	DownloadUrl string `json:"downloadUrl"`
	RepoName    string `json:"repoName"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	Auth        bool   `json:"auth"`
}

type ImageRepoInfo struct {
	ID          uint      `json:"id"`
	CreatedAt   time.Time `json:"createdAt"`
	Name        string    `json:"name"`
	DownloadUrl string    `json:"downloadUrl"`
	RepoName    string    `json:"repoName"`
	Username    string    `json:"username"`
	Auth        bool      `json:"auth"`
}

type ImageRepoOption struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	DownloadUrl string `json:"downloadUrl"`
}
