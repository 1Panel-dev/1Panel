package router

func commonGroups() []CommonRouter {
	return []CommonRouter{
		&VMRouter{},
		&DashboardRouter{},
		&HostRouter{},
		&ContainerRouter{},
		&LogRouter{},
		&FileRouter{},
		&ToolboxRouter{},
		&CronjobRouter{},
		&BackupRouter{},
		&SettingRouter{},
		&AppRouter{},
		&WebsiteRouter{},
		&WebsiteDnsAccountRouter{},
		&WebsiteAcmeAccountRouter{},
		&WebsiteSSLRouter{},
		&WebsiteTemplateRouter{},
		&DatabaseRouter{},
		&NginxRouter{},
		&RuntimeRouter{},
		&ProcessRouter{},
		&WebsiteCARouter{},
		&AIToolsRouter{},
		&GroupRouter{},
		&AlertRouter{},
		&RuntimeDiagnosticsRouter{},
	}
}
