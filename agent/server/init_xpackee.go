//go:build xpackee

package server

import (
	xpack "github.com/1Panel-dev/1Panel/agent/xpack"
)

func InitOthers() {
	xpack.Init()
}
