package common

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/global"
	"gorm.io/gorm"
)

type contextKey string

func initializeTxWatch(db *gorm.DB) {
	_ = db.Callback().Create().Before("gorm:begin_transaction").Register(
		"my_plugin:before_begin",
		func(db *gorm.DB) {
			var caller []string
			for i := 0; ; i++ {
				pc, file, line, ok := runtime.Caller(i)
				if !ok {
					break
				}
				funcName := runtime.FuncForPC(pc).Name()
				if !strings.HasPrefix(funcName, "github.com/1Panel-dev/1Panel") {
					continue
				}
				fileParts := strings.Split(file, "/")
				fileName := fileParts[len(fileParts)-1]
				caller = append(caller, fmt.Sprintf("%s/%s:%d", funcName, fileName, line))
			}
			txID := generateTransactionID()
			db.Statement.Context = context.WithValue(
				db.Statement.Context,
				contextKey("tx_id"), txID,
			)
			db.Statement.Context = context.WithValue(
				db.Statement.Context,
				contextKey("tx_start"), time.Now(),
			)
			global.LOG.Debugf("[%s] tx start \n%s", txID, strings.Join(caller, "\n"))
		},
	)

	_ = db.Callback().Create().After("gorm:commit_or_rollback_transaction").Register(
		"my_plugin:after_commit_or_rollback",
		func(db *gorm.DB) {
			ctx := db.Statement.Context
			txID, _ := ctx.Value(contextKey("tx_id")).(string)
			startTime, _ := ctx.Value(contextKey("tx_start")).(time.Time)
			if txID != "" {
				duration := time.Since(startTime)
				if db.Error != nil {
					global.LOG.Debugf("[%s] tx rollback! time: %v, err: %v", txID, duration, db.Error)
				} else {
					global.LOG.Debugf("[%s] tx commit! time: %v", txID, duration)
				}
			}
		},
	)
}

func generateTransactionID() string {
	return fmt.Sprintf("tx_%d", time.Now().UnixNano())
}
