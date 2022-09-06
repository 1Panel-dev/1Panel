package router

type RouterGroup struct {
	BaseRouter
	UserRouter
	HostRouter
	GroupRouter
	CommandRouter
	MonitorRouter
	OperationLogRouter
}

var RouterGroupApp = new(RouterGroup)
