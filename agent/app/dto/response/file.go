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
	End        bool     `json:"end"`
	Path       string   `json:"path"`
	Total      int      `json:"total"`
	TaskStatus string   `json:"taskStatus"`
	Lines      []string `json:"lines"`
	Scope      string   `json:"scope"`
	TotalLines int      `json:"totalLines"`
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

type FileConvertLog struct {
	Date    string `json:"date"`
	Type    string `json:"type"`
	Log     string `json:"log"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type FileRemarksRes struct {
	Remarks map[string]string `json:"remarks"`
}

type FileAIContentHit struct {
	Path string `json:"path"`
	Line int    `json:"line"`
	Text string `json:"text"`
}

type FileAISearchResult struct {
	Mode                 string             `json:"mode"`
	Summary              string             `json:"summary"`
	Hits                 []FileAIContentHit `json:"hits"`
	ContentScannedFiles  int                `json:"contentScannedFiles"`
	ContentHitsTruncated bool               `json:"contentHitsTruncated"`
	Truncated            bool               `json:"truncated"`
	PreFiltered          bool               `json:"preFiltered"`
	ItemCount            int                `json:"itemCount"`
	PromptTokens         int                `json:"promptTokens"`
	CompletionTokens     int                `json:"completionTokens"`
	TotalTokens          int                `json:"totalTokens"`
	Duration             string             `json:"duration"`
}
