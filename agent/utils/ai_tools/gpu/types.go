package gpu

type Info struct {
	Warnings []string `json:"warnings"`

	CudaVersion   string `json:"cudaVersion"`
	DriverVersion string `json:"driverVersion"`
	Type          string `json:"type"`

	Devices []Device `json:"gpu"`
}

type ECCError struct {
	Scope         string `json:"scope"`
	Correctable   string `json:"correctable"`
	Uncorrectable string `json:"uncorrectable"`
}

type Device struct {
	ECCPending         string     `json:"eccPending"`
	ECCErrors          []ECCError `json:"eccErrors"`
	MemoryActivity     string     `json:"memoryActivity"`
	EncoderUtil        string     `json:"encoderUtil"`
	DecoderUtil        string     `json:"decoderUtil"`
	JPEGUtil           string     `json:"jpegUtil"`
	OFAUtil            string     `json:"ofaUtil"`
	MediaUtil          string     `json:"mediaUtil"`
	HotspotTemperature string     `json:"hotspotTemperature"`
	FanRPM             string     `json:"fanRPM"`
	MediaFrequency     string     `json:"mediaFrequency"`

	UUID              string   `json:"uuid"`
	DriverVersion     string   `json:"driverVersion"`
	Architecture      string   `json:"architecture"`
	Frequency         string   `json:"frequency"`
	MemoryFrequency   string   `json:"memoryFrequency"`
	MemoryTemperature string   `json:"memoryTemperature"`
	MemoryFree        string   `json:"memoryFree"`
	MemoryReserved    string   `json:"memoryReserved"`
	PowerLimit        string   `json:"powerLimit"`
	DefaultPowerLimit string   `json:"defaultPowerLimit"`
	PCIeGeneration    string   `json:"pcieGeneration"`
	PCIeMaxGeneration string   `json:"pcieMaxGeneration"`
	PCIeWidth         string   `json:"pcieWidth"`
	PCIeMaxWidth      string   `json:"pcieMaxWidth"`
	ProcessStatus     string   `json:"processStatus"`
	ClockEvents       []string `json:"clockEvents"`

	Type            string `json:"type"`
	Index           uint   `json:"index"`
	ProductName     string `json:"productName"`
	PersistenceMode string `json:"persistenceMode"`
	BusID           string `json:"busID"`
	DisplayActive   string `json:"displayActive"`
	ECC             string `json:"ecc"`
	FanSpeed        string `json:"fanSpeed"`

	Temperature      string    `json:"temperature"`
	PerformanceState string    `json:"performanceState"`
	PowerDraw        string    `json:"powerDraw"`
	MaxPowerLimit    string    `json:"maxPowerLimit"`
	MemUsed          string    `json:"memUsed"`
	MemTotal         string    `json:"memTotal"`
	GPUUtil          string    `json:"gpuUtil"`
	ComputeMode      string    `json:"computeMode"`
	MigMode          string    `json:"migMode"`
	Processes        []Process `json:"processes"`
}

type Process struct {
	PID         string `json:"pid"`
	Type        string `json:"type"`
	ProcessName string `json:"processName"`
	UsedMemory  string `json:"usedMemory"`
}
