package common

type GpuInfo struct {
	CudaVersion   string `json:"cudaVersion"`
	DriverVersion string `json:"driverVersion"`
	Type          string `json:"type"`

	GPUs []GPU `json:"gpu"`
}

type GPU struct {
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
	Pid         string `json:"pid"`
	Type        string `json:"type"`
	ProcessName string `json:"processName"`
	UsedMemory  string `json:"usedMemory"`
}
