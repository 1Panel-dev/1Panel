package model

import "time"

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

type MonitorVLLM struct {
	ID              uint      `json:"-" gorm:"primarykey;autoIncrement"`
	CreatedAt       time.Time `json:"createdAt"`
	AppInstallID    uint      `json:"appInstallID"`
	Status          string    `json:"status"`
	RawMetrics      string    `json:"-"`
	HistogramDeltas string    `json:"-"`

	Running               *float64 `json:"running"`
	Waiting               *float64 `json:"waiting"`
	CacheUsage            *float64 `json:"cacheUsage"`
	PromptThroughput      *float64 `json:"promptThroughput"`
	GenerationThroughput  *float64 `json:"generationThroughput"`
	RequestThroughput     *float64 `json:"requestThroughput"`
	TimeToFirstToken      *float64 `json:"timeToFirstToken"`
	TimePerOutputToken    *float64 `json:"timePerOutputToken"`
	RequestLatency        *float64 `json:"requestLatency"`
	PrefillTime           *float64 `json:"prefillTime"`
	DecodeTime            *float64 `json:"decodeTime"`
	TimeToFirstTokenP50   *float64 `json:"timeToFirstTokenP50"`
	TimeToFirstTokenP90   *float64 `json:"timeToFirstTokenP90"`
	TimeToFirstTokenP95   *float64 `json:"timeToFirstTokenP95"`
	TimeToFirstTokenP99   *float64 `json:"timeToFirstTokenP99"`
	TimePerOutputTokenP50 *float64 `json:"timePerOutputTokenP50"`
	TimePerOutputTokenP90 *float64 `json:"timePerOutputTokenP90"`
	TimePerOutputTokenP95 *float64 `json:"timePerOutputTokenP95"`
	TimePerOutputTokenP99 *float64 `json:"timePerOutputTokenP99"`
	RequestLatencyP50     *float64 `json:"requestLatencyP50"`
	RequestLatencyP90     *float64 `json:"requestLatencyP90"`
	RequestLatencyP95     *float64 `json:"requestLatencyP95"`
	RequestLatencyP99     *float64 `json:"requestLatencyP99"`
}
