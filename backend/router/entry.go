package router

type RouterGroup struct {
	BaseRouter
	HostRouter
	BackupRouter
	GroupRouter
	CommandRouter
	MonitorRouter
	OperationLogRouter
	FileRouter
	TerminalRouter
	CronjobRouter
	SettingRouter
	AppRouter
}

var RouterGroupApp = new(RouterGroup)
