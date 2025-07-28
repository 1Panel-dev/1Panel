package response

import (
	"time"

	"github.com/1Panel-dev/1Panel/agent/utils/files"
)

type FileInfo struct {
	files.FileInfo
}

type UploadInfo struct {
	Name      string `json:"name"`
	Size      int    `json:"size"`
	CreatedAt string `json:"createdAt"`
}

type FileTree struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Path      string     `json:"path"`
	IsDir     bool       `json:"isDir"`
	Extension string     `json:"extension"`
	Children  []FileTree `json:"children"`
}

type DirSizeRes struct {
	Size int64 `json:"size" validate:"required"`
}

type FileProcessKeys struct {
	Keys []string `json:"keys"`
}

type FileWgetRes struct {
	Key string `json:"key"`
}

type FileLineContent struct {
	Content    string   `json:"content"`
	End        bool     `json:"end"`
	Path       string   `json:"path"`
	Total      int      `json:"total"`
	TaskStatus string   `json:"taskStatus"`
	Lines      []string `json:"lines"`
}

type FileExist struct {
	Exist bool `json:"exist"`
}

type ExistFileInfo struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"modTime"`
	IsDir   bool      `json:"isDir"`
}

type UserInfo struct {
	Username string `json:"username"`
	Group    string `json:"group"`
}

type UserGroupResponse struct {
	Users  []UserInfo `json:"users"`
	Groups []string   `json:"groups"`
}

type DepthDirSizeRes struct {
	Path string `json:"path"`
	Size int64  `json:"size"`
}
