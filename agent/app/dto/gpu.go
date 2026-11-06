package dto

import (
	"time"

	"github.com/1Panel-dev/1Panel/agent/utils/ai_tools/accelerator"
)

type GPURealtimeInfo struct {
	accelerator.Info
	Environment GPUEnvironment `json:"environment"`
	Summaries   []GPUSummary   `json:"summaries"`
}

type GPUEnvironment struct {
	CheckedAt      time.Time             `json:"checkedAt"`
	HardwareStatus string                `json:"hardwareStatus"`
	Checks         []GPUEnvironmentCheck `json:"checks"`
}

type GPUEnvironmentCheck struct {
	Vendor  string `json:"vendor"`
	Item    string `json:"item"`
	Status  string `json:"status"`
	Code    string `json:"code,omitempty"`
	Version string `json:"version,omitempty"`
	Reason  string `json:"reason,omitempty"`
}

type GPUSummary struct {
	Vendor            string   `json:"vendor"`
	DriverVersion     string   `json:"driverVersion"`
	CudaVersion       string   `json:"cudaVersion"`
	CardCount         *int     `json:"cardCount"`
	MemoryTotal       *float64 `json:"memoryTotal"`
	SharedMemoryTotal *float64 `json:"sharedMemoryTotal"`
	MemoryPartial     bool     `json:"memoryPartial"`
}
