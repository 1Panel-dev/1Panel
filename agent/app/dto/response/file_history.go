package response

import "time"

type FileHistoryInfo struct {
	ID             uint      `json:"id"`
	FileID         string    `json:"fileId"`
	Path           string    `json:"path"`
	CurrentPath    string    `json:"currentPath"`
	PreviousID     uint      `json:"previousId"`
	SourcePath     string    `json:"sourcePath"`
	TargetPath     string    `json:"targetPath"`
	FileName       string    `json:"fileName"`
	Extension      string    `json:"extension"`
	FileMode       string    `json:"fileMode"`
	Operation      string    `json:"operation"`
	Deleted        bool      `json:"deleted"`
	ContentSize    int64     `json:"contentSize"`
	ContentSHA     string    `json:"contentSHA"`
	StoragePath    string    `json:"storagePath,omitempty"`
	Content        string    `json:"content,omitempty"`
	CurrentContent string    `json:"currentContent,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type FileHistorySettingInfo struct {
	Enable      string `json:"enable"`
	MaxPerPath  int    `json:"maxPerPath"`
	DiskQuotaMB int    `json:"diskQuotaMB"`
}
