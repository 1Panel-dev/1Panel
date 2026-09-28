package accelerator

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/utils/ai_tools/gpu"
	"github.com/1Panel-dev/1Panel/agent/utils/ai_tools/npu"
	"github.com/1Panel-dev/1Panel/agent/utils/ai_tools/xpu"
)

func normalizeGPU(item *gpu.Device) Device {
	metrics := Metrics{
		MemoryActivity:     metric(item.MemoryActivity, "%"),
		EncoderUtil:        metric(item.EncoderUtil, "%"),
		DecoderUtil:        metric(item.DecoderUtil, "%"),
		JPEGUtil:           metric(item.JPEGUtil, "%"),
		OFAUtil:            metric(item.OFAUtil, "%"),
		MediaUtil:          metric(item.MediaUtil, "%"),
		HotspotTemperature: metric(item.HotspotTemperature, "°C"),
		FanRPM:             metric(item.FanRPM, "RPM"),
		MediaFrequency:     metric(item.MediaFrequency, "MHz"),
		Utilization:        metric(item.GPUUtil, "%"),
		Temperature:        metric(item.Temperature, "°C"),
		MemoryTemperature:  metric(item.MemoryTemperature, "°C"),
		Power:              metric(item.PowerDraw, "W"),
		PowerLimit:         metric(item.PowerLimit, "W"),
		Frequency:          metric(item.Frequency, "MHz"),
		MemoryFrequency:    metric(item.MemoryFrequency, "MHz"),
		MemoryUsed:         memoryMetric(item.MemUsed),
		MemoryTotal:        memoryMetric(item.MemTotal),
		FanSpeed:           metric(item.FanSpeed, "%"),
	}
	device := Device{
		ProcessStatus: item.ProcessStatus,
		ID:            stableID(item.Type, item.UUID, stableID("pci", item.BusID, strconv.FormatUint(uint64(item.Index), 10))),
		Kind:          KindGPU,
		Vendor:        item.Type,
		Index:         int(item.Index),
		Name:          item.ProductName,
		Label:         fmt.Sprintf("%d - %s", item.Index, item.ProductName),
		BusID:         item.BusID,
		Metrics:       metrics,
		GPU:           item,
	}
	for _, process := range item.Processes {
		device.Processes = append(device.Processes, Process{
			PID:    process.PID,
			Type:   process.Type,
			Name:   process.ProcessName,
			Memory: normalizedMemoryDisplay(process.UsedMemory),
		})
	}
	return device
}

func normalizeNPU(item *npu.Device) Device {
	metrics := Metrics{
		AICPUUtil:      metric(item.AICPUUtil, "%"),
		CtrlCPUUtil:    metric(item.CtrlCPUUtil, "%"),
		DDRBandwidth:   metric(item.DDRBandwidth, "%"),
		HBMBandwidth:   metric(item.HBMBandwidth, "%"),
		DDRUsed:        metric(item.MemoryUsed, "MiB"),
		DDRTotal:       metric(item.MemoryTotal, "MiB"),
		HBMUsed:        metric(item.HBMUsed, "MiB"),
		HBMTotal:       metric(item.HBMTotal, "MiB"),
		HugepagesUsed:  metric(item.HugepagesUsed, "pages"),
		HugepagesTotal: metric(item.HugepagesTotal, "pages"),
		Utilization:    metric(item.AICore, "%"),
		Temperature:    metric(item.Temperature, "°C"),
		Power:          metric(item.PowerDraw, "W"),
		MemoryUsed:     memoryMetric(item.MemUsed),
		MemoryTotal:    memoryMetric(item.MemTotal),
	}
	device := Device{
		ProcessStatus: item.ProcessStatus,
		ID:            fmt.Sprintf("%s:%d", stableID("ascend", item.BusID, strconv.FormatUint(uint64(item.NPUIndex), 10)), item.ChipIndex),
		Kind:          KindNPU,
		Vendor:        "ascend",
		Index:         int(item.Index),
		NPUIndex:      int(item.NPUIndex),
		ChipIndex:     int(item.ChipIndex),
		Name:          item.ProductName,
		Label:         fmt.Sprintf("NPU %d / Chip %d - %s", item.NPUIndex, item.ChipIndex, item.ProductName),
		BusID:         item.BusID,
		Metrics:       metrics,
		NPU:           item,
	}
	for _, process := range item.Processes {
		device.Processes = append(device.Processes, Process{
			PID:    process.PID,
			Type:   "NPU",
			Name:   process.ProcessName,
			Memory: normalizedMemoryDisplay(process.UsedMemory),
		})
	}
	return device
}

func normalizeXPU(item *xpu.Device) Device {
	metrics := Metrics{
		MediaUtil:         metric(item.Stats.MediaUtil, "%"),
		ComputeUtil:       metric(item.Stats.ComputeUtil, "%"),
		CopyUtil:          metric(item.Stats.CopyUtil, "%"),
		MediaFrequency:    metric(item.Stats.MediaFrequency, "MHz"),
		MemoryTemperature: metric(item.Stats.MemoryTemperature, "°C"),
		MemoryBandwidth:   metric(item.Stats.MemoryBandwidthUtil, "%"),
		Utilization:       metric(item.Stats.GPUUtil, "%"),
		Temperature:       metric(item.Stats.Temperature, "°C"),
		Power:             metric(item.Stats.Power, "W"),
		MemoryUsed:        memoryMetric(item.Stats.MemoryUsed),
		MemoryTotal:       memoryMetric(item.Basic.Memory),
		MemoryUtil:        metric(item.Stats.MemoryUtil, "%"),
		Frequency:         metric(item.Stats.Frequency, "MHz"),
	}
	device := Device{
		ProcessStatus: item.ProcessStatus,
		ID:            stableID("xpu", item.Basic.UUID, stableID("pci", item.Basic.PciBdfAddress, strconv.Itoa(item.Basic.DeviceID))),
		Kind:          KindXPU,
		Vendor:        item.Basic.VendorName,
		Index:         item.Basic.DeviceID,
		Name:          item.Basic.DeviceName,
		Label:         fmt.Sprintf("%d - %s", item.Basic.DeviceID, item.Basic.DeviceName),
		BusID:         item.Basic.PciBdfAddress,
		Metrics:       metrics,
		XPU:           item,
	}
	for _, process := range item.Processes {
		device.Processes = append(device.Processes, Process{
			PID:          strconv.Itoa(process.PID),
			Type:         process.SHR,
			Name:         process.Command,
			Memory:       normalizedMemoryDisplay(process.Memory),
			SharedMemory: process.SHR,
		})
	}
	return device
}

func stableID(vendor, busID, fallback string) string {
	if busID != "" && !strings.EqualFold(busID, "N/A") {
		return vendor + ":" + busID
	}
	return vendor + ":" + fallback
}
