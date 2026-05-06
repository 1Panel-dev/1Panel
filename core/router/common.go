package router

func commonGroups() []CommonRouter {
	return []CommonRouter{
		&BaseRouter{},
		&UserRouter{},
		&BackupRouter{},
		&LogRouter{},
		&SettingRouter{},
		&CommandRouter{},
		&GroupRouter{},
		&ScriptRouter{},
	}
}
