//go:build xpackee

package server

import (
	xpack "github.com/1Panel-dev/1Panel/core/xpack"
	xpackEE "github.com/1Panel-dev/1Panel/core/xpack-ee"
)

func InitOthers() {
	xpack.Init()
	xpackEE.Init()
}
