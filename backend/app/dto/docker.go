package dto

type DaemonJsonUpdateByFile struct {
	File string `json:"file"`
}

type DaemonJsonConf struct {
	IsSwarm      bool     `json:"isSwarm"`
	Status       string   `json:"status"`
	Version      string   `json:"version"`
	Mirrors      []string `json:"registryMirrors"`
	Registries   []string `json:"insecureRegistries"`
	LiveRestore  bool     `json:"liveRestore"`
	IPTables     bool     `json:"iptables"`
	CgroupDriver string   `json:"cgroupDriver"`

	LogMaxSize string `json:"logMaxSize"`
	LogMaxFile string `json:"logMaxFile"`
}

type LogOption struct {
	LogMaxSize string `json:"logMaxSize"`
	LogMaxFile string `json:"logMaxFile"`
}

type DockerOperation struct {
	Operation string `json:"operation" validate:"required,oneof=start restart stop"`
}
