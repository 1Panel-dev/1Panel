//go:build !xpack && !enterprise

package router

import "github.com/gin-gonic/gin"

func InterceptClusterDatabasePassword(_ *gin.Context, _ string) bool {
	return false
}

func RouterGroups() []CommonRouter {
	return commonGroups()
}

var RouterGroupApp = RouterGroups()
