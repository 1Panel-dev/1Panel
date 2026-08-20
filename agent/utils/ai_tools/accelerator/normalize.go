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
		Utilization: metric(item.GPUUtil, "%"),
		Temperature: metric(item.Temperature, "°C"),
		Power:       metric(item.PowerDraw, "W"),
		PowerLimit:  metric(item.MaxPowerLimit, "W"),
		MemoryUsed:  memoryMetric(item.MemUsed),
		MemoryTotal: memoryMetric(item.MemTotal),
		FanSpeed:    metric(item.FanSpeed, "%"),
	}
	device := Device{
		ID:      stableID(item.Type, item.BusID, strconv.FormatUint(uint64(item.Index), 10)),
		Kind:    KindGPU,
		Vendor:  item.Type,
		Index:   int(item.Index),
		Name:    item.ProductName,
		Label:   fmt.Sprintf("%d - %s", item.Index, item.ProductName),
		BusID:   item.BusID,
		Metrics: metrics,
		GPU:     item,
	}
	device.Capabilities = capabilities(metrics)
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
		Utilization: metric(item.AICore, "%"),
		Temperature: metric(item.Temperature, "°C"),
		Power:       metric(item.PowerDraw, "W"),
		MemoryUsed:  memoryMetric(item.MemUsed),
		MemoryTotal: memoryMetric(item.MemTotal),
	}
	device := Device{
		ID:        fmt.Sprintf("ascend:%d:%d", item.NPUIndex, item.ChipIndex),
		Kind:      KindNPU,
		Vendor:    "ascend",
		Index:     int(item.Index),
		NPUIndex:  int(item.NPUIndex),
		ChipIndex: int(item.ChipIndex),
		Name:      item.ProductName,
		Label:     fmt.Sprintf("NPU %d / Chip %d - %s", item.NPUIndex, item.ChipIndex, item.ProductName),
		BusID:     item.BusID,
		Metrics:   metrics,
		NPU:       item,
	}
	device.Capabilities = capabilities(metrics)
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
		Utilization: metric(item.Stats.GPUUtil, "%"),
		Temperature: metric(item.Stats.Temperature, "°C"),
		Power:       metric(item.Stats.Power, "W"),
		MemoryUsed:  memoryMetric(item.Stats.MemoryUsed),
		MemoryTotal: memoryMetric(item.Basic.Memory),
		MemoryUtil:  metric(item.Stats.MemoryUtil, "%"),
		Frequency:   metric(item.Stats.Frequency, "MHz"),
	}
	device := Device{
		ID:      stableID("xpu", item.Basic.PciBdfAddress, strconv.Itoa(item.Basic.DeviceID)),
		Kind:    KindXPU,
		Vendor:  item.Basic.VendorName,
		Index:   item.Basic.DeviceID,
		Name:    item.Basic.DeviceName,
		Label:   fmt.Sprintf("%d - %s", item.Basic.DeviceID, item.Basic.DeviceName),
		BusID:   item.Basic.PciBdfAddress,
		Metrics: metrics,
		XPU:     item,
	}
	device.Capabilities = capabilities(metrics)
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

func capabilities(metrics Metrics) Capabilities {
	return Capabilities{
		Utilization: metrics.Utilization.Available(),
		Temperature: metrics.Temperature.Available(),
		Power:       metrics.Power.Available(),
		PowerLimit:  metrics.PowerLimit.Available(),
		Memory:      metrics.MemoryUsed.Available() || metrics.MemoryTotal.Available(),
		FanSpeed:    metrics.FanSpeed.Available(),
		Frequency:   metrics.Frequency.Available(),
	}
}

func stableID(vendor, busID, fallback string) string {
	if busID != "" && !strings.EqualFold(busID, "N/A") {
		return vendor + ":" + busID
	}
	return vendor + ":" + fallback
}
