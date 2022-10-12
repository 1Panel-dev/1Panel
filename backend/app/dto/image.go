package dto

import "time"

type ImageInfo struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	Tags      []string  `json:"tags"`
	Size      string    `json:"size"`
}

type ImageLoad struct {
	Path string `josn:"path" validate:"required"`
}

type ImageBuild struct {
	From       string   `josn:"from" validate:"required"`
	Name       string   `json:"name" validate:"required"`
	Dockerfile string   `josn:"dockerfile" validate:"required"`
	Tags       []string `josn:"tags"`
}

type ImagePull struct {
	RepoID    uint   `josn:"repoID"`
	ImageName string `josn:"imageName" validate:"required"`
}

type ImageTag struct {
	RepoID     uint   `josn:"repoID"`
	SourceID   string `json:"sourceID" validate:"required"`
	TargetName string `josn:"targetName" validate:"required"`
}

type ImagePush struct {
	RepoID  uint   `josn:"repoID" validate:"required"`
	TagName string `json:"tagName" validate:"required"`
	Name    string `json:"name" validate:"required"`
}

type ImageSave struct {
	TagName string `json:"tagName" validate:"required"`
	Path    string `josn:"path" validate:"required"`
	Name    string `json:"name" validate:"required"`
}
