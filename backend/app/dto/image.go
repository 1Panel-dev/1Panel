package dto

import "time"

type ImageInfo struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	Name      string    `json:"name"`
	Version   string    `json:"version"`
	Size      string    `json:"size"`
}

type ImageLoad struct {
	Path string `josn:"path" validate:"required"`
}

type ImageRemove struct {
	ImageName string `josn:"imageName" validate:"required"`
}

type ImageBuild struct {
	From       string `josn:"from" validate:"required"`
	Dockerfile string `josn:"dockerfile" validate:"required"`
	Tags       string `josn:"tags" validate:"required"`
}

type ImagePull struct {
	RepoID    uint   `josn:"repoID"`
	ImageName string `josn:"imageName" validate:"required"`
}

type ImagePush struct {
	RepoID    uint   `josn:"repoID" validate:"required"`
	ImageName string `josn:"imageName" validate:"required"`
	TagName   string `json:"tagName" validate:"required"`
}

type ImageSave struct {
	ImageName string `josn:"imageName" validate:"required"`
	Path      string `josn:"path" validate:"required"`
	Name      string `json:"name" validate:"required"`
}
