package schema_v12

import (
	"encoding/xml"

	"github.com/1Panel-dev/1Panel/backend/utils/gpu/common"
)

func Parse(buf []byte) (*common.GpuInfo, error) {
	var (
		s    smi
		info common.GpuInfo
	)
	if err := xml.Unmarshal(buf, &s); err != nil {
		return nil, err
	}

	info.CudaVersion = s.CudaVersion
	info.DriverVersion = s.DriverVersion
	if len(s.Gpu) == 0 {
		return &info, nil
	}
	for i := 0; i < len(s.Gpu); i++ {
		var gpuItem common.GPU
		gpuItem.ProductName = s.Gpu[i].ProductName
		gpuItem.PersistenceMode = s.Gpu[i].PersistenceMode
		gpuItem.BusID = s.Gpu[i].ID
		gpuItem.DisplayActive = s.Gpu[i].DisplayActive
		gpuItem.ECC = s.Gpu[i].EccErrors.Volatile.DramUncorrectable
		gpuItem.FanSpeed = s.Gpu[i].FanSpeed

		gpuItem.Temperature = s.Gpu[i].Temperature.GpuTemp
		gpuItem.PerformanceState = s.Gpu[i].PerformanceState
		gpuItem.PowerDraw = s.Gpu[i].GpuPowerReadings.PowerDraw
		gpuItem.MaxPowerLimit = s.Gpu[i].GpuPowerReadings.MaxPowerLimit
		gpuItem.MemUsed = s.Gpu[i].FbMemoryUsage.Used
		gpuItem.MemTotal = s.Gpu[i].FbMemoryUsage.Total
		gpuItem.GPUUtil = s.Gpu[i].Utilization.GpuUtil
		gpuItem.ComputeMode = s.Gpu[i].ComputeMode
		gpuItem.MigMode = s.Gpu[i].MigMode.CurrentMig

		for _, process := range s.Gpu[i].Processes.ProcessInfo {
			gpuItem.Processes = append(gpuItem.Processes, common.Process{
				Pid:         process.Pid,
				Type:        process.Type,
				ProcessName: process.ProcessName,
				UsedMemory:  process.UsedMemory,
			})
		}
		info.Gpu = append(info.Gpu, gpuItem)
	}
	return &info, nil
}
