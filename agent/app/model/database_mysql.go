package model

type DatabaseMysql struct {
	BaseModel
	Name        string `json:"name" gorm:"not null"`
	From        string `json:"from" gorm:"not null;default:local"`
	MysqlName   string `json:"mysqlName" gorm:"not null"`
	Format      string `json:"format" gorm:"not null"`
	Username    string `json:"username" gorm:"not null"`
	Password    string `json:"password" gorm:"not null"`
	Permission  string `json:"permission" gorm:"not null"`
	IsDelete    bool   `json:"isDelete"`
	Description string `json:"description"`
}
