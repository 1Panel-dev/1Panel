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
	CgroupDriver string   `json:"cgroupDriver"`
}

type DockerOperation struct {
	Operation string `json:"operation" validate:"required,oneof=start restart stop"`
}
