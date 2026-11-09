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

type MonitorGPU struct {
	MemoryUtil         *float64 `json:"memoryUtil"`
	MemoryActivity     *float64 `json:"memoryActivity"`
	EncoderUtil        *float64 `json:"encoderUtil"`
	DecoderUtil        *float64 `json:"decoderUtil"`
	JPEGUtil           *float64 `json:"jpegUtil"`
	OFAUtil            *float64 `json:"ofaUtil"`
	MediaUtil          *float64 `json:"mediaUtil"`
	ComputeUtil        *float64 `json:"computeUtil"`
	CopyUtil           *float64 `json:"copyUtil"`
	HotspotTemperature *float64 `json:"hotspotTemperature"`
	FanRPM             *float64 `json:"fanRPM"`
	AICPUUtil          *float64 `json:"aiCPUUtil"`
	CtrlCPUUtil        *float64 `json:"ctrlCPUUtil"`
	DDRUsed            *float64 `json:"ddrUsed"`
	DDRTotal           *float64 `json:"ddrTotal"`
	HBMUsed            *float64 `json:"hbmUsed"`
	HBMTotal           *float64 `json:"hbmTotal"`
	DDRBandwidth       *float64 `json:"ddrBandwidth"`
	HBMBandwidth       *float64 `json:"hbmBandwidth"`
	MemoryBandwidth    *float64 `json:"memoryBandwidth"`
	MediaFrequency     *float64 `json:"mediaFrequency"`
	HugepagesUsed      *float64 `json:"hugepagesUsed"`
	HugepagesTotal     *float64 `json:"hugepagesTotal"`

	MemoryTemperature *float64 `json:"memoryTemperature"`
	DeviceID          string   `json:"deviceID"`
	DeviceType        string   `json:"deviceType"`
	ProcessStatus     string   `json:"processStatus"`
	Frequency         *float64 `json:"frequency"`
	MemoryFrequency   *float64 `json:"memoryFrequency"`
	IntervalSeconds   int      `json:"intervalSeconds"`
	BaseModel
	ProductName   string   `json:"productName"`
	GPUUtil       *float64 `json:"gpuUtil"`
	Temperature   *float64 `json:"temperature"`
	PowerDraw     *float64 `json:"powerDraw"`
	MaxPowerLimit *float64 `json:"maxPowerLimit"`
	MemUsed       *float64 `json:"memUsed"`
	MemTotal      *float64 `json:"memTotal"`
	FanSpeed      *float64 `json:"fanSpeed"`
	Processes     string   `json:"processes"`
}
