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

type MonitorGPUOptions struct {
	GPUType   string         `json:"gpuType"`
	ChartHide []GPUChartHide `json:"chartHide"`
	Options   []string       `json:"options"`
}
type GPUChartHide struct {
	ProductName string `json:"productName"`
	Process     bool   `json:"process"`
	GPU         bool   `json:"gpu"`
	Memory      bool   `json:"memory"`
	Power       bool   `json:"power"`
	Temperature bool   `json:"temperature"`
	Speed       bool   `json:"speed"`
}
type MonitorGPUSearch struct {
	ProductName string    `json:"productName"`
	StartTime   time.Time `json:"startTime"`
	EndTime     time.Time `json:"endTime"`
}
type MonitorGPUData struct {
	Date             []time.Time `json:"date"`
	GPUValue         []float64   `json:"gpuValue"`
	TemperatureValue []float64   `json:"temperatureValue"`
	PowerTotal       []float64   `json:"powerTotal"`
	PowerUsed        []float64   `json:"powerUsed"`
	PowerPercent     []float64   `json:"powerPercent"`
	MemoryTotal      []float64   `json:"memoryTotal"`
	MemoryUsed       []float64   `json:"memoryUsed"`
	MemoryPercent    []float64   `json:"memoryPercent"`
	SpeedValue       []int       `json:"speedValue"`

	ProcessCount []int          `json:"processCount"`
	GPUProcesses [][]GPUProcess `json:"gpuProcesses"`
}

type GPUProcess struct {
	Pid         string `json:"pid"`
	Type        string `json:"type"`
	ProcessName string `json:"processName"`
	UsedMemory  string `json:"usedMemory"`
}
