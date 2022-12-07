package dto

type DaemonJsonUpdateByFile struct {
	Path string `json:"path" validate:"required"`
	File string `json:"file"`
}

type DaemonJsonConf struct {
	Status       string   `json:"status"`
	Mirrors      []string `json:"registryMirrors"`
	Registries   []string `json:"insecureRegistries"`
	Bip          string   `json:"bip"`
	LiveRestore  bool     `json:"liveRestore"`
	CgroupDriver string   `json:"cgroupDriver"`
}

type DockerOperation struct {
	Operation string `json:"operation" validate:"required,oneof=start restart stop"`
}
