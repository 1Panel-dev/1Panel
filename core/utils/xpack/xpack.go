//go:build !xpack

package xpack

import (
	"io"

	"github.com/gin-gonic/gin"
)

func Proxy(c *gin.Context, currentNode string) {}

func UpdateGroup(name string, group, newGroup uint) error { return nil }

func CheckBackupUsed(id uint) error { return nil }

func RequestToAllAgent(reqUrl, reqMethod string, reqBody io.Reader) error { return nil }
