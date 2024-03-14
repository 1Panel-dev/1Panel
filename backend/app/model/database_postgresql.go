package model

type DatabasePostgresql struct {
	BaseModel
	Name           string `json:"name" gorm:"type:varchar(256);not null"`
	From           string `json:"from" gorm:"type:varchar(256);not null;default:local"`
	PostgresqlName string `json:"postgresqlName" gorm:"type:varchar(64);not null"`
	Format         string `json:"format" gorm:"type:varchar(64);not null"`
	Username       string `json:"username" gorm:"type:varchar(256);not null"`
	Password       string `json:"password" gorm:"type:varchar(256);not null"`
	SuperUser      bool   `json:"superUser" gorm:"type:varchar(64)"`
	IsDelete       bool   `json:"isDelete" gorm:"type:varchar(64)"`
	Description    string `json:"description" gorm:"type:varchar(256);"`
}
