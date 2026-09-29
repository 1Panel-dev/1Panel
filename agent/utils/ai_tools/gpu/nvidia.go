package gpu

import (
	"context"
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

const nvidiaSMICommand = "nvidia-smi"

type nvidiaSMI struct{}

func (n nvidiaSMI) LoadInfo(ctx context.Context) (*Info, error) {
	cmdMgr := cmd.NewCommandMgr(cmd.WithContext(ctx), cmd.WithTimeout(5*time.Second))
	itemData, err := cmdMgr.RunWithStdout(nvidiaSMICommand, "-q", "-x")
	if err != nil {
		return nil, fmt.Errorf("calling %s failed: %w", nvidiaSMICommand, err)
	}
	return parseNvidiaSMI([]byte(itemData))
}

func parseNvidiaSMI(buf []byte) (*Info, error) {
	var (
		s    nvidiaSMIResponse
		info Info
	)
	if err := xml.Unmarshal(buf, &s); err != nil {
		return nil, err
	}

	info.Type = "nvidia"
	info.CudaVersion = s.CudaVersion
	info.DriverVersion = s.DriverVersion
	for i := range s.Gpu {
		gpuItem := Device{
			Type:              "nvidia",
			MemoryActivity:    s.Gpu[i].Utilization.MemoryUtil,
			EncoderUtil:       s.Gpu[i].Utilization.EncoderUtil,
			DecoderUtil:       s.Gpu[i].Utilization.DecoderUtil,
			JPEGUtil:          s.Gpu[i].Utilization.JpegUtil,
			OFAUtil:           s.Gpu[i].Utilization.OfaUtil,
			MediaFrequency:    s.Gpu[i].Clocks.VideoClock,
			UUID:              s.Gpu[i].UUID,
			DriverVersion:     s.DriverVersion,
			Architecture:      s.Gpu[i].ProductArchitecture,
			Frequency:         s.Gpu[i].Clocks.GraphicsClock,
			MemoryFrequency:   s.Gpu[i].Clocks.MemClock,
			MemoryTemperature: s.Gpu[i].Temperature.MemoryTemp,
			MemoryFree:        s.Gpu[i].FbMemoryUsage.Free,
			MemoryReserved:    s.Gpu[i].FbMemoryUsage.Reserved,
			PCIeGeneration:    firstSMIValue(s.Gpu[i].Pci.PciGpuLinkInfo.PcieGen.DeviceCurrentLinkGen, s.Gpu[i].Pci.PciGpuLinkInfo.PcieGen.CurrentLinkGen),
			PCIeMaxGeneration: firstSMIValue(s.Gpu[i].Pci.PciGpuLinkInfo.PcieGen.MaxDeviceLinkGen, s.Gpu[i].Pci.PciGpuLinkInfo.PcieGen.MaxLinkGen),
			PCIeWidth:         s.Gpu[i].Pci.PciGpuLinkInfo.LinkWidths.CurrentLinkWidth,
			PCIeMaxWidth:      s.Gpu[i].Pci.PciGpuLinkInfo.LinkWidths.MaxLinkWidth,
			ProcessStatus:     "unavailable",
			Index:             uint(i),
			ProductName:       s.Gpu[i].ProductName,
			PersistenceMode:   s.Gpu[i].PersistenceMode,
			BusID:             s.Gpu[i].ID,
			DisplayActive:     s.Gpu[i].DisplayActive,
			ECC:               s.Gpu[i].EccMode.CurrentEcc,
			FanSpeed:          s.Gpu[i].FanSpeed,
			Temperature:       s.Gpu[i].Temperature.GpuTemp,
			PerformanceState:  s.Gpu[i].PerformanceState,
			MemUsed:           s.Gpu[i].FbMemoryUsage.Used,
			MemTotal:          s.Gpu[i].FbMemoryUsage.Total,
			GPUUtil:           s.Gpu[i].Utilization.GpuUtil,
			ComputeMode:       s.Gpu[i].ComputeMode,
			MigMode:           s.Gpu[i].MigMode.CurrentMig,
		}
		gpuItem.ECCPending = s.Gpu[i].EccMode.PendingEcc
		gpuItem.ECCErrors = []ECCError{
			{Scope: "Volatile Total", Correctable: s.Gpu[i].EccErrors.Volatile.SingleBit.Total, Uncorrectable: s.Gpu[i].EccErrors.Volatile.DoubleBit.Total},
			{Scope: "Aggregate Total", Correctable: s.Gpu[i].EccErrors.Aggregate.SingleBit.Total, Uncorrectable: s.Gpu[i].EccErrors.Aggregate.DoubleBit.Total},

			{Scope: "Volatile DRAM", Correctable: s.Gpu[i].EccErrors.Volatile.DramCorrectable, Uncorrectable: s.Gpu[i].EccErrors.Volatile.DramUncorrectable},
			{Scope: "Volatile SRAM", Correctable: s.Gpu[i].EccErrors.Volatile.SramCorrectable, Uncorrectable: s.Gpu[i].EccErrors.Volatile.SramUncorrectable},
			{Scope: "Aggregate DRAM", Correctable: s.Gpu[i].EccErrors.Aggregate.DramCorrectable, Uncorrectable: s.Gpu[i].EccErrors.Aggregate.DramUncorrectable},
			{Scope: "Aggregate SRAM", Correctable: s.Gpu[i].EccErrors.Aggregate.SramCorrectable, Uncorrectable: s.Gpu[i].EccErrors.Aggregate.SramUncorrectable},
		}
		gpuItem.PowerDraw = firstSMIValue(s.Gpu[i].GpuPowerReadings.PowerDraw, s.Gpu[i].GpuPowerReadings.InstantPowerDraw, s.Gpu[i].PowerReadings.PowerDraw)
		gpuItem.PowerLimit = firstSMIValue(s.Gpu[i].GpuPowerReadings.CurrentPowerLimit, s.Gpu[i].PowerReadings.EnforcedPowerLimit, s.Gpu[i].PowerReadings.PowerLimit)
		gpuItem.MaxPowerLimit = firstSMIValue(s.Gpu[i].GpuPowerReadings.MaxPowerLimit, s.Gpu[i].PowerReadings.MaxPowerLimit)
		gpuItem.DefaultPowerLimit = firstSMIValue(s.Gpu[i].GpuPowerReadings.DefaultPowerLimit, s.Gpu[i].PowerReadings.DefaultPowerLimit)
		for _, event := range append(s.Gpu[i].ClocksEventReasons.Reasons, s.Gpu[i].ClocksThrottleReasons.Reasons...) {
			if strings.EqualFold(strings.TrimSpace(event.Value), "Active") {
				gpuItem.ClockEvents = append(gpuItem.ClockEvents, strings.TrimPrefix(strings.TrimPrefix(event.XMLName.Local, "clocks_event_reason_"), "clocks_throttle_reason_"))
			}
		}
		if s.Gpu[i].Processes != nil && strings.TrimSpace(s.Gpu[i].Processes.Text) == "" {
			gpuItem.ProcessStatus = "ok"
		}
		if s.Gpu[i].Processes != nil {
			for _, process := range s.Gpu[i].Processes.ProcessInfo {
				gpuItem.Processes = append(gpuItem.Processes, Process{
					PID:         process.Pid,
					Type:        process.Type,
					ProcessName: process.ProcessName,
					UsedMemory:  process.UsedMemory,
				})
			}
		}
		info.Devices = append(info.Devices, gpuItem)
	}
	return &info, nil
}

func firstSMIValue(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" && !strings.EqualFold(value, "N/A") && !strings.EqualFold(value, "Not Supported") {
			return value
		}
	}
	return ""
}
