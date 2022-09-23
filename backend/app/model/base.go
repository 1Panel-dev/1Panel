package model

import "time"

type BaseModel struct {
	ID        uint `gorm:"primarykey;AUTO_INCREMENT"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
