package dto

import "time"

type ImageInfo struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	IsUsed    bool      `json:"isUsed"`
	Tags      []string  `json:"tags"`
	Size      string    `json:"size"`
}

type ImageLoad struct {
	Path string `json:"path" validate:"required"`
}

type ImageBuild struct {
	From       string   `json:"from" validate:"required"`
	Name       string   `json:"name" validate:"required"`
	Dockerfile string   `json:"dockerfile" validate:"required"`
	Tags       []string `json:"tags"`
}

type ImagePull struct {
	RepoID    uint   `json:"repoID"`
	ImageName string `json:"imageName" validate:"required"`
}

type ImageTag struct {
	SourceID   string `json:"sourceID" validate:"required"`
	TargetName string `json:"targetName" validate:"required"`
}

type ImagePush struct {
	RepoID  uint   `json:"repoID" validate:"required"`
	TagName string `json:"tagName" validate:"required"`
	Name    string `json:"name" validate:"required"`
}

type ImageSave struct {
	TagName string `json:"tagName" validate:"required"`
	Path    string `json:"path" validate:"required"`
	Name    string `json:"name" validate:"required"`
}
