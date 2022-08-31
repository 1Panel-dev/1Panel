package router

type RouterGroup struct {
	BaseRouter
	UserRouter
	HostRouter
	GroupRouter
	CommandRouter
	OperationLogRouter
}

var RouterGroupApp = new(RouterGroup)
