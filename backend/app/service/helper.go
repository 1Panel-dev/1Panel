package service

import (
	"context"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"gorm.io/gorm"
)

func getTxAndContext() (tx *gorm.DB, ctx context.Context) {
	tx = global.DB.Begin()
	ctx = context.WithValue(context.Background(), constant.DB, tx)
	return
}

func getTxByContext(ctx context.Context) (*gorm.DB, context.Context) {
	tx, ok := ctx.Value(constant.DB).(*gorm.DB)
	if ok {
		return tx, ctx
	}
	tx = global.DB.Begin()
	ctx = context.WithValue(context.Background(), constant.DB, tx)
	return tx, ctx
}
