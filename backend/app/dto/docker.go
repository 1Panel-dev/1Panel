package dto

type DaemonJsonUpdateByFile struct {
	File string `json:"file"`
}

type DaemonJsonConf struct {
	Status       string   `json:"status"`
	Version      string   `json:"version"`
	Mirrors      []string `json:"registryMirrors"`
	Registries   []string `json:"insecureRegistries"`
	LiveRestore  bool     `json:"liveRestore"`
	IPTables     bool     `json:"iptables"`
	CgroupDriver string   `json:"cgroupDriver"`
}

type DockerOperation struct {
	StopSocket  bool   `json:"stopSocket"`
	StopService bool   `json:"stopService"`
	Operation   string `json:"operation" validate:"required,oneof=start restart stop"`
}
