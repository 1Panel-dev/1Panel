//go:build !xpack

package xpack

import (
	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/gin-gonic/gin"
)

func Proxy(c *gin.Context, currentNode string) error {
	return nil
}

type Node struct{}

func SyncBackupOperation(operate string, accounts []model.BackupAccount) error {
	return nil
}
