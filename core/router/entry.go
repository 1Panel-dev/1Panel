//go:build !xpack

package router

func RouterGroups() []CommonRouter {
	return commonGroups()
}

var RouterGroupApp = RouterGroups()
