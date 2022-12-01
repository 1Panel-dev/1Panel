package dto

import (
	"github.com/1Panel-dev/1Panel/backend/utils/files"
)

type FileOption struct {
	files.FileOption
}

type FileInfo struct {
	files.FileInfo
}

type FileTree struct {
	ID       string     `json:"id"`
	Name     string     `json:"name"`
	Path     string     `json:"path"`
	Children []FileTree `json:"children"`
}

type FileCreate struct {
	Path      string
	Content   string
	IsDir     bool
	Mode      int64
	IsLink    bool
	IsSymlink bool
	LinkPath  string
}

type FileDelete struct {
	Path  string
	IsDir bool
}

type FileBatchDelete struct {
	IsDir bool
	Paths []string
}

type FileCompress struct {
	Files   []string
	Dst     string
	Type    string
	Name    string
	Replace bool
}

type FileDeCompress struct {
	Dst  string
	Type string
	Path string
}

type FileEdit struct {
	Path    string
	Content string
}

type FileRename struct {
	OldName string
	NewName string
}

type FileWget struct {
	Url  string `json:"url" validate:"required"`
	Path string `json:"path" validate:"required"`
	Name string `json:"name" validate:"required"`
}

type FileMove struct {
	Type     string   `json:"type" validate:"required"`
	OldPaths []string `json:"oldPaths" validate:"required"`
	NewPath  string   `json:"newPath" validate:"required"`
}

type FileDownload struct {
	Paths []string `json:"paths" validate:"required"`
	Type  string   `json:"type" validate:"required"`
	Name  string   `json:"name" validate:"required"`
}

type DirSizeReq struct {
	Path string `json:"path" validate:"required"`
}

type DirSizeRes struct {
	Size float64 `json:"size" validate:"required"`
}

type FileProcess struct {
	Total   uint64  `json:"total"`
	Written uint64  `json:"written"`
	Percent float64 `json:"percent"`
	Name    string  `json:"name"`
}

type FileProcessReq struct {
	Key string `json:"key"`
}

type FileProcessKeys struct {
	Keys []string `json:"keys"`
}

type FileWgetRes struct {
	Key string
}
