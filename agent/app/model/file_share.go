package model

type FileShare struct {
	BaseModel
	Path          string `gorm:"not null;uniqueIndex" json:"path"`
	Token         string `gorm:"not null;uniqueIndex" json:"token"`
	FileName      string `gorm:"not null" json:"fileName"`
	ExpiresUnix   int64  `json:"expiresUnix"`
	PasswordEnc   string `json:"-"`
	PasswordSalt  string `json:"passwordSalt"`
	PasswordHash  string `json:"passwordHash"`
	MaxDownloads  int    `json:"maxDownloads"`
	DownloadCount int    `json:"downloadCount"`
}
