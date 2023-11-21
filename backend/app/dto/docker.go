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

	Ipv6         bool   `json:"ipv6"`
	FixedCidrV6  string `json:"fixedCidrV6"`
	Ip6Tables    bool   `json:"ip6Tables"`
	Experimental bool   `json:"experimental"`

	LogMaxSize string `json:"logMaxSize"`
	LogMaxFile string `json:"logMaxFile"`
}

type LogOption struct {
	LogMaxSize string `json:"logMaxSize"`
	LogMaxFile string `json:"logMaxFile"`
}

type Ipv6Option struct {
	FixedCidrV6  string `json:"fixedCidrV6"`
	Ip6Tables    bool   `json:"ip6Tables" validate:"required"`
	Experimental bool   `json:"experimental"`
}

type DockerOperation struct {
	Operation string `json:"operation" validate:"required,oneof=start restart stop"`
}
