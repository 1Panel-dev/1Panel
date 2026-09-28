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
	DeviceID    string `json:"deviceID"`
	Legacy      bool   `json:"legacy"`
	ProductName string `json:"productName"`
	Type        string `json:"type"`
	Process     bool   `json:"process"`
	GPU         bool   `json:"gpu"`
	Memory      bool   `json:"memory"`
	Power       bool   `json:"power"`
	PowerLimit  bool   `json:"powerLimit"`
	Temperature bool   `json:"temperature"`
	Speed       bool   `json:"speed"`
}
type MonitorGPUSearch struct {
	Aggregation string    `json:"aggregation" validate:"omitempty,oneof=avg max"`
	DeviceID    string    `json:"deviceID"`
	Legacy      bool      `json:"legacy"`
	ProductName string    `json:"productName"`
	StartTime   time.Time `json:"startTime"`
	EndTime     time.Time `json:"endTime"`
}
type MonitorGPUData struct {
	MemoryActivity     []*float64 `json:"memoryActivity"`
	EncoderUtil        []*float64 `json:"encoderUtil"`
	DecoderUtil        []*float64 `json:"decoderUtil"`
	JPEGUtil           []*float64 `json:"jpegUtil"`
	OFAUtil            []*float64 `json:"ofaUtil"`
	MediaUtil          []*float64 `json:"mediaUtil"`
	ComputeUtil        []*float64 `json:"computeUtil"`
	CopyUtil           []*float64 `json:"copyUtil"`
	HotspotTemperature []*float64 `json:"hotspotTemperature"`
	FanRPM             []*float64 `json:"fanRPM"`
	AICPUUtil          []*float64 `json:"aiCPUUtil"`
	CtrlCPUUtil        []*float64 `json:"ctrlCPUUtil"`
	DDRUsed            []*float64 `json:"ddrUsed"`
	DDRTotal           []*float64 `json:"ddrTotal"`
	HBMUsed            []*float64 `json:"hbmUsed"`
	HBMTotal           []*float64 `json:"hbmTotal"`
	DDRBandwidth       []*float64 `json:"ddrBandwidth"`
	HBMBandwidth       []*float64 `json:"hbmBandwidth"`
	MemoryBandwidth    []*float64 `json:"memoryBandwidth"`
	MediaFrequency     []*float64 `json:"mediaFrequency"`
	HugepagesUsed      []*float64 `json:"hugepagesUsed"`
	HugepagesTotal     []*float64 `json:"hugepagesTotal"`

	BucketSeconds          int64       `json:"bucketSeconds"`
	SampleCount            int64       `json:"sampleCount"`
	MemoryTemperatureValue []*float64  `json:"memoryTemperatureValue"`
	FrequencyValue         []*float64  `json:"frequencyValue"`
	MemoryFrequencyValue   []*float64  `json:"memoryFrequencyValue"`
	Date                   []time.Time `json:"date"`
	GPUValue               []*float64  `json:"gpuValue"`
	TemperatureValue       []*float64  `json:"temperatureValue"`
	PowerTotal             []*float64  `json:"powerTotal"`
	PowerUsed              []*float64  `json:"powerUsed"`
	PowerPercent           []*float64  `json:"powerPercent"`
	MemoryTotal            []*float64  `json:"memoryTotal"`
	MemoryUsed             []*float64  `json:"memoryUsed"`
	MemoryPercent          []*float64  `json:"memoryPercent"`
	SpeedValue             []*float64  `json:"speedValue"`

	ProcessCount []*float64     `json:"processCount"`
	GPUProcesses [][]GPUProcess `json:"gpuProcesses"`
}

type GPUProcess struct {
	Pid         string `json:"pid"`
	Type        string `json:"type"`
	ProcessName string `json:"processName"`
	UsedMemory  string `json:"usedMemory"`
}
