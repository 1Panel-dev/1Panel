package service

import (
	"context"

	"github.com/1Panel-dev/1Panel/backend/global"
	"gorm.io/gorm"
)

type dbStr string

func getTxAndContext() (tx *gorm.DB, ctx context.Context) {
	db := dbStr("db")
	tx = global.DB.Begin()
	ctx = context.WithValue(context.Background(), db, tx)
	return
}
