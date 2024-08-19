//go:build !xpack

package xpack

import (
	"github.com/gin-gonic/gin"
)

func Proxy(c *gin.Context, currentNode string) error {
	return nil
}

func UpdateGroup(name string, group, newGroup uint) error {
	return nil
}

func CheckBackupUsed(id uint) error {
	return nil
}
