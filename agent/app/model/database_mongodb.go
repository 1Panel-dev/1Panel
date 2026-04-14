package model

type DatabaseMongodb struct {
	BaseModel
	Name        string `json:"name" gorm:"not null"`
	From        string `json:"from" gorm:"not null;default:local"`
	MongodbName string `json:"mongodbName" gorm:"not null"`
	Username    string `json:"username" gorm:"not null"`
	Password    string `json:"password" gorm:"not null"`
	IsDelete    bool   `json:"isDelete"`
	Description string `json:"description"`
}
