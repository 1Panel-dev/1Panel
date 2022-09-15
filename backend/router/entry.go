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
	SettingRouter
}

var RouterGroupApp = new(RouterGroup)
