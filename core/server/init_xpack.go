//go:build xpack

package server

import (
	xpack "github.com/1Panel-dev/1Panel/core/xpack"
)

func InitOthers() {
	xpack.Init()
}
