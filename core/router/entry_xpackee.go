//go:build xpackee

package router

import (
	xpackEERouter "github.com/1Panel-dev/1Panel/core/xpack-ee/router"
	xpackRouter "github.com/1Panel-dev/1Panel/core/xpack/router"
)

func RouterGroups() []CommonRouter {
	baseRouter := commonGroups()
	for _, ro := range xpackRouter.XpackGroups() {
		if val, ok := ro.(CommonRouter); ok {
			baseRouter = append(baseRouter, val)
		}
	}
	for _, ro := range xpackEERouter.XpackEEGroups() {
		if val, ok := ro.(CommonRouter); ok {
			baseRouter = append(baseRouter, val)
		}
	}
	return baseRouter
}

var RouterGroupApp = RouterGroups()
