package router

func commonGroups() []CommonRouter {
	return []CommonRouter{
		&BaseRouter{},
		&LogRouter{},
		&SettingRouter{},
	}
}
