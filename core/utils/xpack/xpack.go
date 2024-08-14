//go:build !xpack

package xpack

import (
	"github.com/gin-gonic/gin"
)

func Proxy(c *gin.Context, currentNode string) error {
	return nil
}
