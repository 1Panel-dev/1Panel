package gpu

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/global"
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
	data := []byte(itemData)
	version := "v11"

	buf := bytes.NewBuffer(data)
	decoder := xml.NewDecoder(buf)
	for {
		token, err := decoder.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("reading token failed: %w", err)
		}
		d, ok := token.(xml.Directive)
		if !ok {
			continue
		}
		directive := string(d)
		if !strings.HasPrefix(directive, "DOCTYPE") {
			continue
		}
		parts := strings.Split(directive, " ")
		s := strings.Trim(parts[len(parts)-1], "\" ")
		if strings.HasPrefix(s, "nvsmi_device_") && strings.HasSuffix(s, ".dtd") {
			version = strings.TrimSuffix(strings.TrimPrefix(s, "nvsmi_device_"), ".dtd")
		} else {
			global.LOG.Debugf("Cannot find schema version in %q", directive)
		}
		break
	}

	return parseNvidiaSMI(data, version)
}

func parseNvidiaSMI(buf []byte, version string) (*Info, error) {
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
			Type:             "nvidia",
			Index:            uint(i),
			ProductName:      s.Gpu[i].ProductName,
			PersistenceMode:  s.Gpu[i].PersistenceMode,
			BusID:            s.Gpu[i].ID,
			DisplayActive:    s.Gpu[i].DisplayActive,
			ECC:              s.Gpu[i].EccErrors.Volatile.DramUncorrectable,
			FanSpeed:         s.Gpu[i].FanSpeed,
			Temperature:      s.Gpu[i].Temperature.GpuTemp,
			PerformanceState: s.Gpu[i].PerformanceState,
			MemUsed:          s.Gpu[i].FbMemoryUsage.Used,
			MemTotal:         s.Gpu[i].FbMemoryUsage.Total,
			GPUUtil:          s.Gpu[i].Utilization.GpuUtil,
			ComputeMode:      s.Gpu[i].ComputeMode,
			MigMode:          s.Gpu[i].MigMode.CurrentMig,
		}
		if version == "v12" || version == "v13" {
			gpuItem.PowerDraw = s.Gpu[i].GpuPowerReadings.PowerDraw
			if gpuItem.PowerDraw == "" {
				gpuItem.PowerDraw = s.Gpu[i].GpuPowerReadings.InstantPowerDraw
			}
			gpuItem.MaxPowerLimit = s.Gpu[i].GpuPowerReadings.CurrentPowerLimit
		} else {
			gpuItem.PowerDraw = s.Gpu[i].PowerReadings.PowerDraw
			gpuItem.MaxPowerLimit = s.Gpu[i].PowerReadings.MaxPowerLimit
		}

		for _, process := range s.Gpu[i].Processes.ProcessInfo {
			gpuItem.Processes = append(gpuItem.Processes, Process{
				PID:         process.Pid,
				Type:        process.Type,
				ProcessName: process.ProcessName,
				UsedMemory:  process.UsedMemory,
			})
		}
		info.Devices = append(info.Devices, gpuItem)
	}
	return &info, nil
}
