package model

type DatabaseUser struct {
	BaseModel
	Type        string `json:"type" gorm:"not null;uniqueIndex:idx_database_user"`
	Database    string `json:"database" gorm:"not null;uniqueIndex:idx_database_user"`
	Username    string `json:"username" gorm:"not null;uniqueIndex:idx_database_user"`
	Host        string `json:"host" gorm:"uniqueIndex:idx_database_user"`
	Password    string `json:"password"`
	Description string `json:"description"`
	IsDelete    bool   `json:"isDelete"`
}
