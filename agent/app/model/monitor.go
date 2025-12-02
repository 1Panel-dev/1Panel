package model

type MonitorBase struct {
	BaseModel
	Cpu         float64     `json:"cpu"`
	TopCPU      string      `json:"topCPU"`
	TopCPUItems interface{} `gorm:"-" json:"topCPUItems"`

	Memory      float64     `json:"memory"`
	TopMem      string      `json:"topMem"`
	TopMemItems interface{} `gorm:"-" json:"topMemItems"`

	LoadUsage float64 `json:"loadUsage"`
	CpuLoad1  float64 `json:"cpuLoad1"`
	CpuLoad5  float64 `json:"cpuLoad5"`
	CpuLoad15 float64 `json:"cpuLoad15"`
}

type MonitorIO struct {
	BaseModel
	Name  string `json:"name"`
	Read  uint64 `json:"read"`
	Write uint64 `json:"write"`
	Count uint64 `json:"count"`
	Time  uint64 `json:"time"`
}

type MonitorNetwork struct {
	BaseModel
	Name string  `json:"name"`
	Up   float64 `json:"up"`
	Down float64 `json:"down"`
}
