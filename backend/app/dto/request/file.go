package request

import (
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
)

type FileOption struct {
	files.FileOption
}

type FileContentReq struct {
	Path     string `json:"path" validate:"required"`
	IsDetail bool   `json:"isDetail"`
}

type SearchUploadWithPage struct {
	dto.PageInfo
	Path string `json:"path" validate:"required"`
}

type FileCreate struct {
	Path      string `json:"path" validate:"required"`
	Content   string `json:"content"`
	IsDir     bool   `json:"isDir"`
	Mode      int64  `json:"mode"`
	IsLink    bool   `json:"isLink"`
	IsSymlink bool   `json:"isSymlink"`
	LinkPath  string `json:"linkPath"`
	Sub       bool   `json:"sub"`
}

type FileRoleReq struct {
	Paths []string `json:"paths" validate:"required"`
	Mode  int64    `json:"mode" validate:"required"`
	User  string   `json:"user" validate:"required"`
	Group string   `json:"group" validate:"required"`
	Sub   bool     `json:"sub"`
}

type FileDelete struct {
	Path        string `json:"path" validate:"required"`
	IsDir       bool   `json:"isDir"`
	ForceDelete bool   `json:"forceDelete"`
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
	Secret  string   `json:"secret"`
}

type FileDeCompress struct {
	Dst    string `json:"dst"  validate:"required"`
	Type   string `json:"type"  validate:"required"`
	Path   string `json:"path" validate:"required"`
	Secret string `json:"secret"`
}

type FileEdit struct {
	Path    string `json:"path"  validate:"required"`
	Content string `json:"content"`
}

type FileRename struct {
	OldName string `json:"oldName" validate:"required"`
	NewName string `json:"newName" validate:"required"`
}

type FilePathCheck struct {
	Path string `json:"path" validate:"required"`
}

type FileWget struct {
	Url               string `json:"url" validate:"required"`
	Path              string `json:"path" validate:"required"`
	Name              string `json:"name" validate:"required"`
	IgnoreCertificate bool   `json:"ignoreCertificate"`
}

type FileMove struct {
	Type     string   `json:"type" validate:"required"`
	OldPaths []string `json:"oldPaths" validate:"required"`
	NewPath  string   `json:"newPath" validate:"required"`
	Name     string   `json:"name"`
	Cover    bool     `json:"cover"`
}

type FileDownload struct {
	Paths    []string `json:"paths" validate:"required"`
	Type     string   `json:"type" validate:"required"`
	Name     string   `json:"name" validate:"required"`
	Compress bool     `json:"compress"`
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
	Sub   bool   `json:"sub"`
}

type FileReadByLineReq struct {
	Page     int    `json:"page" validate:"required"`
	PageSize int    `json:"pageSize" validate:"required"`
	Type     string `json:"type" validate:"required"`
	ID       uint   `json:"ID"`
	Name     string `json:"name"`
	Latest   bool   `json:"latest"`
}

type FileExistReq struct {
	Name string `json:"name" validate:"required"`
	Dir  string `json:"dir" validate:"required"`
}
