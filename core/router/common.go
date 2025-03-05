package router

func commonGroups() []CommonRouter {
	return []CommonRouter{
		&BaseRouter{},
		&BackupRouter{},
		&LogRouter{},
		&SettingRouter{},
		&CommandRouter{},
		&HostRouter{},
		&GroupRouter{},
		&AppLauncherRouter{},
		&ScriptRouter{},
	}
}
