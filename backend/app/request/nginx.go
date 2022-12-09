package request

type NginxConfigFileUpdate struct {
	Content  string `json:"content" validate:"required"`
	FilePath string `json:"filePath" validate:"required"`
	Backup   bool   `json:"backup" validate:"required"`
}
