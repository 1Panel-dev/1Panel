package dto

import (
	"github.com/1Panel-dev/1Panel/utils/files"
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
	Path    string
	Content string
	IsDir   bool
	Mode    int64
}
