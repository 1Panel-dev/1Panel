package router

type RouterGroup struct {
	BaseRouter
	UserRouter
	HostRouter
	GroupRouter
	CommandRouter
	OperationLogRouter
	FileRouter
}

var RouterGroupApp = new(RouterGroup)
