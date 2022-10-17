package service

import (
	"context"
	"github.com/1Panel-dev/1Panel/backend/global"
	"gorm.io/gorm"
)

func getTxAndContext() (tx *gorm.DB, ctx context.Context) {
	tx = global.DB.Begin()
	ctx = context.WithValue(context.Background(), "db", tx)
	return
}
