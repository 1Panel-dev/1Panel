//go:build xpack

package server

import "github.com/1Panel-dev/1Panel/backend/xpack"

func InitOthers() {
	xpack.InitXpack()
}
