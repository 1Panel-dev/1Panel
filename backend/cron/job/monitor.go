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
		var aheadData model.MonitorIO
		if err := global.DB.Where("name = ?", v.Name).Order("created_at").Find(&aheadData).Error; err != nil {
			_ = global.DB.Create(&itemIO)
			continue
		}
		stime := time.Since(aheadData.CreatedAt).Seconds()
		itemIO.Read = uint64(float64(v.ReadBytes-aheadData.ReadByte) / stime)
		itemIO.Write = uint64(float64(v.WriteBytes-aheadData.WriteByte) / stime)

		itemIO.Count = uint64(float64(v.ReadCount-aheadData.ReadCount) / stime)
		writeCount := uint64(float64(v.WriteCount-aheadData.WriteCount) / stime)
		if writeCount > itemIO.Count {
			itemIO.Count = writeCount
		}

		itemIO.Time = uint64(float64(v.ReadTime-aheadData.ReadTime) / stime)
		writeTime := uint64(float64(v.WriteTime-aheadData.WriteTime) / stime)
		if writeTime > itemIO.Time {
			itemIO.Time = writeTime
		}
		_ = global.DB.Create(&itemIO)
	}

	netStat, _ := net.IOCounters(true)
	for _, v := range netStat {
		var itemNet model.MonitorNetwork
		var aheadData model.MonitorNetwork
		itemNet.Name = v.Name
		itemNet.BytesSent = v.BytesSent
		itemNet.BytesRecv = v.BytesRecv
		if err := global.DB.Where("name = ?", v.Name).Order("created_at").Find(&aheadData).Error; err != nil {
			_ = global.DB.Create(&itemNet)
			continue
		}
		stime := time.Since(aheadData.CreatedAt).Seconds()
		itemNet.Up = float64(v.BytesSent-aheadData.BytesSent) / 1024 / stime
		itemNet.Down = float64(v.BytesRecv-aheadData.BytesRecv) / 1024 / stime
		_ = global.DB.Create(&itemNet)
	}
}
