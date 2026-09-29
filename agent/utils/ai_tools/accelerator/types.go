package accelerator

import (
	"errors"
	"time"

	"github.com/1Panel-dev/1Panel/agent/utils/ai_tools/gpu"
	"github.com/1Panel-dev/1Panel/agent/utils/ai_tools/npu"
	"github.com/1Panel-dev/1Panel/agent/utils/ai_tools/xpu"
)

type Kind string

const (
	KindGPU Kind = "gpu"
	KindNPU Kind = "npu"
	KindXPU Kind = "xpu"
)

type Info struct {
	CollectedAt time.Time `json:"collectedAt"`
	Warnings    []string  `json:"warnings"`

	Type             string       `json:"type"`
	CudaVersion      string       `json:"cudaVersion"`
	DriverVersion    string       `json:"driverVersion"`
	XPUDriverVersion string       `json:"xpuDriverVersion"`
	GPUs             []gpu.Device `json:"gpu"`
	NPUs             []npu.Device `json:"npu"`
	XPUs             []xpu.Device `json:"xpu"`
}

type Metric struct {
	Value   *float64
	Unit    string
	Display string
}

func (m Metric) Available() bool {
	return m.Value != nil
}

func (m Metric) ValueOrZero() float64 {
	if m.Value == nil {
		return 0
	}
	return *m.Value
}

type Metrics struct {
	MemoryActivity     Metric
	EncoderUtil        Metric
	DecoderUtil        Metric
	JPEGUtil           Metric
	OFAUtil            Metric
	MediaUtil          Metric
	ComputeUtil        Metric
	CopyUtil           Metric
	HotspotTemperature Metric
	FanRPM             Metric
	AICPUUtil          Metric
	CtrlCPUUtil        Metric
	DDRUsed            Metric
	DDRTotal           Metric
	HBMUsed            Metric
	HBMTotal           Metric
	DDRBandwidth       Metric
	HBMBandwidth       Metric
	MemoryBandwidth    Metric
	MediaFrequency     Metric
	HugepagesUsed      Metric
	HugepagesTotal     Metric

	MemoryTemperature Metric
	Utilization       Metric
	Temperature       Metric
	Power             Metric
	PowerLimit        Metric
	MemoryUsed        Metric
	MemoryTotal       Metric
	MemoryUtil        Metric
	FanSpeed          Metric
	Frequency         Metric
	MemoryFrequency   Metric
}

type Process struct {
	PID          string
	Type         string
	Name         string
	Memory       string
	SharedMemory string
}

type Device struct {
	ParentID      string
	ProcessStatus string
	ID            string
	Kind          Kind
	Vendor        string
	Index         int
	NPUIndex      int
	ChipIndex     int
	Name          string
	Label         string
	BusID         string
	Metrics       Metrics
	Processes     []Process

	GPU *gpu.Device `json:"-"`
	NPU *npu.Device `json:"-"`
	XPU *xpu.Device `json:"-"`
}

type Snapshot struct {
	Info     Info
	Devices  []Device
	Warnings []error
}

func (s Snapshot) Warning() error {
	return errors.Join(s.Warnings...)
}

type ProviderSnapshot struct {
	Warnings      []string
	Type          string
	DriverVersion string
	CudaVersion   string
	GPUs          []gpu.Device
	NPUs          []npu.Device
	XPUs          []xpu.Device
	Devices       []Device
}
