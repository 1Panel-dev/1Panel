package response

type HostToolRes struct {
	Type   string      `json:"type"`
	Config interface{} `json:"config"`
}

type Supervisor struct {
	ConfigPath  string `json:"configPath"`
	IncludeDir  string `json:"includeDir"`
	LogPath     string `json:"logPath"`
	IsExist     bool   `json:"isExist"`
	Init        bool   `json:"init"`
	Msg         string `json:"msg"`
	Version     string `json:"version"`
	Status      string `json:"status"`
	CtlExist    bool   `json:"ctlExist"`
	ServiceName string `json:"serviceName"`
}

type HostToolConfig struct {
	Content string `json:"content"`
}

type SupervisorProcessConfig struct {
	Name     string          `json:"name"`
	Command  string          `json:"command"`
	User     string          `json:"user"`
	Dir      string          `json:"dir"`
	Numprocs string          `json:"numprocs"`
	Msg      string          `json:"msg"`
	Status   []ProcessStatus `json:"status"`
}

type ProcessStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	PID    string `json:"PID"`
	Uptime string `json:"uptime"`
	Msg    string `json:"msg"`
}
