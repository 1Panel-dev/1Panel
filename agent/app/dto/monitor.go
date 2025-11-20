package dto

import "time"

type MonitorSearch struct {
	Param     string    `json:"param" validate:"required,oneof=all cpu memory load io network"`
	IO        string    `json:"io"`
	Network   string    `json:"network"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

type MonitorData struct {
	Param string        `json:"param"`
	Date  []time.Time   `json:"date"`
	Value []interface{} `json:"value"`
}

type Process struct {
	Name    string  `json:"name"`
	Pid     int32   `json:"pid"`
	Percent float64 `json:"percent"`
	Memory  uint64  `json:"memory"`
	Cmd     string  `json:"cmd"`
	User    string  `json:"user"`
}

type MonitorSetting struct {
	MonitorStatus    string `json:"monitorStatus"`
	MonitorStoreDays string `json:"monitorStoreDays"`
	MonitorInterval  string `json:"monitorInterval"`
	DefaultNetwork   string `json:"defaultNetwork"`
	DefaultIO        string `json:"defaultIO"`
}

type MonitorSettingUpdate struct {
	Key   string `json:"key" validate:"required,oneof=MonitorStatus MonitorStoreDays MonitorInterval DefaultNetwork DefaultIO"`
	Value string `json:"value"`
}

type MonitorGPUSearch struct {
	ProductName string    `json:"productName"`
	StartTime   time.Time `json:"startTime"`
	EndTime     time.Time `json:"endTime"`
}
type MonitorGPUData struct {
	ProductNames     []string               `json:"productNames"`
	Date             []time.Time            `json:"date"`
	GPUValue         []float64              `json:"gpuValue"`
	TemperatureValue []int                  `json:"temperatureValue"`
	PowerValue       []GPUPowerUsageHelper  `json:"powerValue"`
	MemoryValue      []GPUMemoryUsageHelper `json:"memoryValue"`
	SpeedValue       []int                  `json:"speedValue"`
}
type GPUPowerUsageHelper struct {
	Total   float64 `json:"total"`
	Used    float64 `json:"used"`
	Percent float64 `json:"percent"`
}
type GPUMemoryUsageHelper struct {
	Total   int     `json:"total"`
	Used    int     `json:"used"`
	Percent float64 `json:"percent"`

	GPUProcesses []GPUProcess `json:"gpuProcesses"`
}
type GPUProcess struct {
	Pid         string `json:"pid"`
	Type        string `json:"type"`
	ProcessName string `json:"processName"`
	UsedMemory  string `json:"usedMemory"`
}
