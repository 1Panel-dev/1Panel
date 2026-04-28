//go:build !xpack && !enterprise

package router

func RouterGroups() []CommonRouter {
	return commonGroups()
}

var RouterGroupApp = RouterGroups()
