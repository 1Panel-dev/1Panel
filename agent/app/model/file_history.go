package model

type FileHistory struct {
	BaseModel
	FileID      string `json:"fileId" gorm:"not null;index:idx_file_history_file_id_created_at,priority:1;index:idx_file_history_file_id_deleted_created_at,priority:1"`
	Path        string `json:"path" gorm:"not null;index:idx_file_history_path_created_at,priority:1;index:idx_file_history_path_hash_created_at,priority:1;index:idx_file_history_related_path,priority:1"`
	PathHash    string `json:"pathHash" gorm:"not null;index:idx_file_history_path_created_at,priority:2;index:idx_file_history_path_hash_created_at,priority:2"`
	PreviousID  uint   `json:"previousId" gorm:"index"`
	SourcePath  string `json:"sourcePath" gorm:"index:idx_file_history_related_path,priority:2"`
	TargetPath  string `json:"targetPath" gorm:"index:idx_file_history_related_path,priority:3"`
	FileName    string `json:"fileName" gorm:"not null"`
	Extension   string `json:"extension"`
	FileMode    string `json:"fileMode"`
	Operation   string `json:"operation" gorm:"not null;index:idx_file_history_operation_created_at,priority:1"`
	Deleted     bool   `json:"deleted" gorm:"not null;default:false;index:idx_file_history_deleted_created_at,priority:1;index:idx_file_history_file_id_deleted_created_at,priority:2"`
	ContentSize int64  `json:"contentSize" gorm:"not null"`
	ContentSHA  string `json:"contentSHA" gorm:"not null;index"`
	StoragePath string `json:"storagePath" gorm:"not null;uniqueIndex"`
}
