//go:build !xpack && !xpackee

package router

func RouterGroups() []CommonRouter {
	return commonGroups()
}

var RouterGroupApp = RouterGroups()
