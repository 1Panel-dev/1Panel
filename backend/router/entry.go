package router

type RouterGroup struct {
	BaseRouter
	UserRouter
	HostRouter
	GroupRouter
	CommandRouter
	MonitorRouter
	OperationLogRouter
	FileRouter
}

var RouterGroupApp = new(RouterGroup)
