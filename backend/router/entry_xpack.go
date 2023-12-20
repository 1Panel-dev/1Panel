//go:build xpack

package router

import "github.com/1Panel-dev/1Panel/backend/xpack/router"

func RouterGroups() []CommonRouter {
	return append(commonGroups(), []CommonRouter{&router.WafRouter{}}...)
}

var RouterGroupApp = RouterGroups()
