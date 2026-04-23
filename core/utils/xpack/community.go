//go:build !xpack && !xpackee

package xpack

import "github.com/1Panel-dev/1Panel/core/utils/xpack/helper"

var AuthProvider = helper.NewIAuthProvider()

var MultiNodeProvider = helper.NewIMultiNodeProvider()
