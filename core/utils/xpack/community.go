//go:build !xpack && !enterprise

package xpack

import (
	"github.com/1Panel-dev/1Panel/core/utils/xpack/helper"
	"github.com/gin-gonic/gin"
)

var AuthProvider = helper.NewIAuthProvider()

var MultiNodeProvider = helper.NewIMultiNodeProvider()

func VMCountLimit(_ *gin.Context) int {
	return 1
}
