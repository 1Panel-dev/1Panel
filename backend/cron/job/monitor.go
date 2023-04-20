package job

import (
	"strconv"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/app/repo"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

type monitor struct{}

func NewMonitorJob() *monitor {
	return &monitor{}
}

func (m *monitor) Run() {
	settingRepo := repo.NewISettingRepo()
	monitorStatus, _ := settingRepo.Get(settingRepo.WithByKey("MonitorStatus"))
	if monitorStatus.Value == "disable" {
		return
	}
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

	if err := global.DB.Create(&itemModel).Error; err != nil {
		global.LOG.Errorf("Insert basic monitoring data failed, err: %v", err)
	}

	go loadDiskIO()
	go loadNetIO()

	MonitorStoreDays, err := settingRepo.Get(settingRepo.WithByKey("MonitorStoreDays"))
	if err != nil {
		return
	}
	storeDays, _ := strconv.Atoi(MonitorStoreDays.Value)
	timeForDelete := time.Now().AddDate(0, 0, -storeDays)
	_ = global.DB.Where("created_at < ?", timeForDelete).Delete(&model.MonitorBase{}).Error
	_ = global.DB.Where("created_at < ?", timeForDelete).Delete(&model.MonitorIO{}).Error
	_ = global.DB.Where("created_at < ?", timeForDelete).Delete(&model.MonitorNetwork{}).Error
}

func loadDiskIO() {
	ioStat, _ := disk.IOCounters()

	time.Sleep(60 * time.Second)

	ioStat2, _ := disk.IOCounters()
	var ioList []model.MonitorIO
	for _, io2 := range ioStat2 {
		for _, io1 := range ioStat {
			if io2.Name == io1.Name {
				var itemIO model.MonitorIO
				itemIO.Name = io1.Name
				itemIO.Read = uint64(float64(io2.ReadBytes-io1.ReadBytes) / 60)
				itemIO.Write = uint64(float64(io2.WriteBytes-io1.WriteBytes) / 60)

				itemIO.Count = uint64(float64(io2.ReadCount-io1.ReadCount) / 60)
				writeCount := uint64(float64(io2.WriteCount-io1.WriteCount) / 60)
				if writeCount > itemIO.Count {
					itemIO.Count = writeCount
				}

				itemIO.Time = uint64(float64(io2.ReadTime-io1.ReadTime) / 60)
				writeTime := uint64(float64(io2.WriteTime-io1.WriteTime) / 60)
				if writeTime > itemIO.Time {
					itemIO.Time = writeTime
				}
				ioList = append(ioList, itemIO)
				break
			}
		}
	}
	if err := global.DB.CreateInBatches(ioList, len(ioList)).Error; err != nil {
		global.LOG.Errorf("Insert io monitoring data failed, err: %v", err)
	}
}

func loadNetIO() {
	netStat, _ := net.IOCounters(true)
	netStatAll, _ := net.IOCounters(false)

	time.Sleep(60 * time.Second)

	netStat2, _ := net.IOCounters(true)
	var netList []model.MonitorNetwork
	for _, net2 := range netStat2 {
		for _, net1 := range netStat {
			if net2.Name == net1.Name {
				var itemNet model.MonitorNetwork
				itemNet.Name = net1.Name
				itemNet.Up = float64(net2.BytesSent-net1.BytesSent) / 1024 / 60
				itemNet.Down = float64(net2.BytesRecv-net1.BytesRecv) / 1024 / 60
				netList = append(netList, itemNet)
				break
			}
		}
	}
	netStatAll2, _ := net.IOCounters(false)
	for _, net2 := range netStatAll2 {
		for _, net1 := range netStatAll {
			if net1.BytesSent == 0 || net1.BytesRecv == 0 {
				continue
			}
			if net2.Name == net1.Name {
				var itemNet model.MonitorNetwork
				itemNet.Name = net1.Name
				itemNet.Up = float64(net2.BytesSent-net1.BytesSent) / 1024 / 60
				itemNet.Down = float64(net2.BytesRecv-net1.BytesRecv) / 1024 / 60
				netList = append(netList, itemNet)
				break
			}
		}
	}

	if err := global.DB.CreateInBatches(netList, len(netList)).Error; err != nil {
		global.LOG.Errorf("Insert network monitoring data failed, err: %v", err)
	}
}
