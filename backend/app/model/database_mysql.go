package model

type DatabaseMysql struct {
	BaseModel
	Name          string `json:"name" gorm:"type:varchar(256);not null"`
	Format        string `json:"format" gorm:"type:varchar(64);not null"`
	Username      string `json:"username" gorm:"type:varchar(256);not null"`
	Password      string `json:"password" gorm:"type:varchar(256);not null"`
	Permission    string `json:"permission" gorm:"type:varchar(256);not null"`
	PermissionIPs string `json:"permissionIPs" gorm:"type:varchar(256)"`
	Description   string `json:"description" gorm:"type:varchar(256);"`
}
