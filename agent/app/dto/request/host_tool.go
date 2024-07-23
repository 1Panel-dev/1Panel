package request

type HostToolReq struct {
	Type    string `json:"type" validate:"required,oneof=supervisord"`
	Operate string `json:"operate" validate:"oneof=status restart start stop"`
}

type HostToolCreate struct {
	Type string `json:"type" validate:"required"`
	SupervisorConfig
}

type SupervisorConfig struct {
	ConfigPath  string `json:"configPath"`
	ServiceName string `json:"serviceName"`
}

type HostToolLogReq struct {
	Type string `json:"type" validate:"required,oneof=supervisord"`
}

type HostToolConfig struct {
	Type    string `json:"type" validate:"required,oneof=supervisord"`
	Operate string `json:"operate" validate:"oneof=get set"`
	Content string `json:"content"`
}

type SupervisorProcessConfig struct {
	Name     string `json:"name"`
	Operate  string `json:"operate"`
	Command  string `json:"command"`
	User     string `json:"user"`
	Dir      string `json:"dir"`
	Numprocs string `json:"numprocs"`
}
type SupervisorProcessFileReq struct {
	Name    string `json:"name" validate:"required"`
	Operate string `json:"operate" validate:"required,oneof=get clear update" `
	Content string `json:"content"`
	File    string `json:"file" validate:"required,oneof=out.log err.log config"`
}
