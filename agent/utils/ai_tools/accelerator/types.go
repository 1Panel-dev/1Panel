package accelerator

import (
	"errors"

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
	Utilization Metric
	Temperature Metric
	Power       Metric
	PowerLimit  Metric
	MemoryUsed  Metric
	MemoryTotal Metric
	MemoryUtil  Metric
	FanSpeed    Metric
	Frequency   Metric
}

type Capabilities struct {
	Utilization bool
	Temperature bool
	Power       bool
	PowerLimit  bool
	Memory      bool
	FanSpeed    bool
	Frequency   bool
}

type Process struct {
	PID          string
	Type         string
	Name         string
	Memory       string
	SharedMemory string
}

type Device struct {
	ID           string
	Kind         Kind
	Vendor       string
	Index        int
	NPUIndex     int
	ChipIndex    int
	Name         string
	Label        string
	BusID        string
	Metrics      Metrics
	Capabilities Capabilities
	Processes    []Process

	GPU *gpu.Device `json:"-"`
	NPU *npu.Device `json:"-"`
	XPU *xpu.Device `json:"-"`
}

type Snapshot struct {
	Info           Info
	Devices        []Device
	DriverVersions map[string]string
	Warnings       []error
}

func (s Snapshot) Warning() error {
	return errors.Join(s.Warnings...)
}

type ProviderSnapshot struct {
	Type          string
	DriverVersion string
	CudaVersion   string
	GPUs          []gpu.Device
	NPUs          []npu.Device
	XPUs          []xpu.Device
	Devices       []Device
}
