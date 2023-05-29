package request

import (
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
)

type FileOption struct {
	files.FileOption
}

type SearchUploadWithPage struct {
	dto.PageInfo
	Path string `json:"path" validate:"required"`
}

type FileCreate struct {
	Path      string `json:"path" validate:"required"`
	Content   string `json:"content"`
	IsDir     bool   `json:"isDir"`
	Mode      int64  `json:"mode" validate:"required"`
	IsLink    bool   `json:"isLink"`
	IsSymlink bool   `json:"isSymlink"`
	LinkPath  string `json:"linkPath"`
	Sub       bool   `json:"sub"`
}

type FileDelete struct {
	Path  string `json:"path" validate:"required"`
	IsDir bool   `json:"isDir"`
}

type FileBatchDelete struct {
	Paths []string `json:"paths" validate:"required"`
	IsDir bool     `json:"isDir"`
}

type FileCompress struct {
	Files   []string `json:"files" validate:"required"`
	Dst     string   `json:"dst" validate:"required"`
	Type    string   `json:"type" validate:"required"`
	Name    string   `json:"name" validate:"required"`
	Replace bool     `json:"replace"`
}

type FileDeCompress struct {
	Dst  string `json:"dst"  validate:"required"`
	Type string `json:"type"  validate:"required"`
	Path string `json:"path" validate:"required"`
}

type FileEdit struct {
	Path    string `json:"path"  validate:"required"`
	Content string `json:"content"  validate:"required"`
}

type FileRename struct {
	OldName string `json:"oldName" validate:"required"`
	NewName string `json:"newName" validate:"required"`
}

type FilePathCheck struct {
	Path string `json:"path" validate:"required"`
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
	Paths    []string `json:"paths" validate:"required"`
	Type     string   `json:"type" validate:"required"`
	Name     string   `json:"name" validate:"required"`
	Compress bool     `json:"compress" validate:"required"`
}

type FileChunkDownload struct {
	Path string `json:"path" validate:"required"`
	Name string `json:"name" validate:"required"`
}

type DirSizeReq struct {
	Path string `json:"path" validate:"required"`
}

type FileProcessReq struct {
	Key string `json:"key"`
}

type FileRoleUpdate struct {
	Path  string `json:"path" validate:"required"`
	User  string `json:"user" validate:"required"`
	Group string `json:"group" validate:"required"`
	Sub   bool   `json:"sub" validate:"required"`
}
