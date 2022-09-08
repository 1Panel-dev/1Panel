package router

type RouterGroup struct {
	BaseRouter
	HostRouter
	GroupRouter
	CommandRouter
	MonitorRouter
	OperationLogRouter
	FileRouter
	TerminalRouter
	SettingRouter
}

var RouterGroupApp = new(RouterGroup)
