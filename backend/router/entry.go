package router

type RouterGroup struct {
	BaseRouter
	HostRouter
	BackupRouter
	GroupRouter
	ContainerRouter
	CommandRouter
	MonitorRouter
	OperationLogRouter
	FileRouter
	TerminalRouter
	CronjobRouter
	SettingRouter
	AppRouter
	WebsiteRouter
	WebsiteGroupRouter
	WebsiteDnsAccountRouter
	WebsiteAcmeAccountRouter
	DatabaseRouter
}

var RouterGroupApp = new(RouterGroup)
