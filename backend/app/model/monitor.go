package model

type MonitorBase struct {
	BaseModel
	Cpu float64 `gorm:"type:float" json:"cpu"`

	LoadUsage float64 `gorm:"type:float" json:"loadUsage"`
	CpuLoad1  float64 `gorm:"type:float" json:"cpuLoad1"`
	CpuLoad5  float64 `gorm:"type:float" json:"cpuLoad5"`
	CpuLoad15 float64 `gorm:"type:float" json:"cpuLoad15"`

	Memory float64 `gorm:"type:float" json:"memory"`
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
	Up   float64 `gorm:"type:float" json:"up"`
	Down float64 `gorm:"type:float" json:"down"`
}
