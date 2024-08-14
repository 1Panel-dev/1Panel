package model

type DatabasePostgresql struct {
	BaseModel
	Name           string `json:"name" gorm:"not null"`
	From           string `json:"from" gorm:"not null;default:local"`
	PostgresqlName string `json:"postgresqlName" gorm:"not null"`
	Format         string `json:"format" gorm:"not null"`
	Username       string `json:"username" gorm:"not null"`
	Password       string `json:"password" gorm:"not null"`
	SuperUser      bool   `json:"superUser"`
	IsDelete       bool   `json:"isDelete"`
	Description    string `json:"description"`
}
