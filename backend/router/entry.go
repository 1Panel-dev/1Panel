package router

type RouterGroup struct {
	BaseRouter
	DashboardRouter
	HostRouter
	ContainerRouter
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
	RuntimeRouter
}

var RouterGroupApp = new(RouterGroup)
