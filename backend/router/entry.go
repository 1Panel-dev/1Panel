package router

type RouterGroup struct {
	BaseRouter
	DashboardRouter
	HostRouter
	BackupRouter
	GroupRouter
	ContainerRouter
	CommandRouter
	MonitorRouter
	LogRouter
	FileRouter
	TerminalRouter
	CronjobRouter
	SettingRouter
	AppRouter
	WebsiteRouter
	WebsiteGroupRouter
	WebsiteDnsAccountRouter
	WebsiteAcmeAccountRouter
	WebsiteSSLRouter
	DatabaseRouter
	NginxRouter
}

var RouterGroupApp = new(RouterGroup)
