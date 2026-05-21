//go:build !xpack && !enterprise

package xpack

import "github.com/1Panel-dev/1Panel/agent/utils/xpack/helper"

var AlertProvider = helper.NewIAlertProvider()

var MultiNodeProvider = helper.NewIMultiNodeProvider()
