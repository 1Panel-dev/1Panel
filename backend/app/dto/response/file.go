package response

import "github.com/1Panel-dev/1Panel/backend/utils/files"

type FileInfo struct {
	files.FileInfo
}

type FileTree struct {
	ID       string     `json:"id"`
	Name     string     `json:"name"`
	Path     string     `json:"path"`
	Children []FileTree `json:"children"`
}

type DirSizeRes struct {
	Size float64 `json:"size" validate:"required"`
}

type FileProcessKeys struct {
	Keys []string `json:"keys"`
}

type FileWgetRes struct {
	Key string `json:"key"`
}
