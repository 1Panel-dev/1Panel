package job

import (
	"time"

	"github.com/1Panel-dev/1Panel/app/model"
	"github.com/1Panel-dev/1Panel/global"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
)

type monitor struct{}

func NewMonitorJob() *monitor {
	return &monitor{}
}

func (m *monitor) Run() {
	var itemModel model.MonitorBase
	totalPercent, _ := cpu.Percent(3*time.Second, false)
	if len(totalPercent) == 1 {
		itemModel.Cpu = totalPercent[0]
	}
	cpuCount, _ := cpu.Counts(false)

	loadInfo, _ := load.Avg()
	itemModel.CpuLoad1 = loadInfo.Load1
	itemModel.CpuLoad5 = loadInfo.Load5
	itemModel.CpuLoad15 = loadInfo.Load15
	itemModel.LoadUsage = loadInfo.Load1 / (float64(cpuCount*2) * 0.75) * 100

	memoryInfo, _ := mem.VirtualMemory()
	itemModel.Memory = memoryInfo.UsedPercent

	_ = global.DB.Create(&itemModel)

	ioStat, _ := disk.IOCounters()
	for _, v := range ioStat {
		var itemIO model.MonitorIO
		itemIO.Name = v.Name
		itemIO.ReadCount = v.ReadCount
		itemIO.WriteCount = v.WriteCount
		itemIO.ReadByte = v.ReadBytes
		itemIO.WriteByte = v.WriteBytes
		itemIO.ReadTime = v.ReadTime
		itemIO.WriteTime = v.WriteTime
		_ = global.DB.Create(&itemIO)
	}

	netStat, _ := net.IOCounters(true)
	for _, v := range netStat {
		var itemNet model.MonitorNetwork
		itemNet.Name = v.Name
		itemNet.BytesSent = v.BytesSent
		itemNet.BytesRecv = v.BytesRecv
		_ = global.DB.Create(&itemNet)
	}
}
