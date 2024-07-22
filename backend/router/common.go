package router

func commonGroups() []CommonRouter {
	return []CommonRouter{
		&DashboardRouter{},
		&HostRouter{},
		&ContainerRouter{},
		&LogRouter{},
		&FileRouter{},
		&ToolboxRouter{},
		&TerminalRouter{},
		&CronjobRouter{},
		&SettingRouter{},
		&AppRouter{},
		&WebsiteRouter{},
		&WebsiteGroupRouter{},
		&WebsiteDnsAccountRouter{},
		&WebsiteAcmeAccountRouter{},
		&WebsiteSSLRouter{},
		&DatabaseRouter{},
		&NginxRouter{},
		&RuntimeRouter{},
		&ProcessRouter{},
		&WebsiteCARouter{},
	}
}
