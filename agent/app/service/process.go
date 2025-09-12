package service

import (
	"fmt"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/websocket"
	"github.com/shirou/gopsutil/v4/process"
	"time"
)

type ProcessService struct{}

type IProcessService interface {
	StopProcess(req request.ProcessReq) error
	GetProcessInfoByPID(pid int32) (*websocket.PsProcessData, error)
}

func NewIProcessService() IProcessService {
	return &ProcessService{}
}

func (ps *ProcessService) StopProcess(req request.ProcessReq) error {
	proc, err := process.NewProcess(req.PID)
	if err != nil {
		return err
	}
	if err := proc.Kill(); err != nil {
		return err
	}
	return nil
}

func (ps *ProcessService) GetProcessInfoByPID(pid int32) (*websocket.PsProcessData, error) {
	p, err := process.NewProcess(pid)
	if err != nil {
		return nil, fmt.Errorf("get process info by pid %v: %v", pid, err)
	}

	exists, err := p.IsRunning()
	if err != nil || !exists {
		return nil, fmt.Errorf("process %v is not running", pid)
	}

	data := &websocket.PsProcessData{
		PID: pid,
	}

	if name, err := p.Name(); err == nil {
		data.Name = name
	}

	if ppid, err := p.Ppid(); err == nil {
		data.PPID = ppid
	}

	if username, err := p.Username(); err == nil {
		data.Username = username
	}

	if status, err := p.Status(); err == nil {
		if len(status) > 0 {
			data.Status = status[0]
		}
	}

	if createTime, err := p.CreateTime(); err == nil {
		data.StartTime = time.Unix(createTime/1000, 0).Format("2006-01-02 15:04:05")
	}

	if numThreads, err := p.NumThreads(); err == nil {
		data.NumThreads = numThreads
	}

	if connections, err := p.Connections(); err == nil {
		data.NumConnections = len(connections)

		var connects []websocket.ProcessConnect
		for _, conn := range connections {
			pc := websocket.ProcessConnect{
				Status: conn.Status,
				Laddr:  conn.Laddr,
				Raddr:  conn.Raddr,
				PID:    pid,
				Name:   data.Name,
			}
			connects = append(connects, pc)
		}
		data.Connects = connects
	}

	if cpuPercent, err := p.CPUPercent(); err == nil {
		data.CpuValue = cpuPercent
		data.CpuPercent = fmt.Sprintf("%.2f%%", cpuPercent)
	}

	if ioCounters, err := p.IOCounters(); err == nil {
		data.DiskRead = common.FormatBytes(ioCounters.ReadBytes)
		data.DiskWrite = common.FormatBytes(ioCounters.WriteBytes)
	}

	if cmdline, err := p.Cmdline(); err == nil {
		data.CmdLine = cmdline
	}

	if memInfo, err := p.MemoryInfo(); err == nil {
		data.Rss = common.FormatBytes(memInfo.RSS)
		data.VMS = common.FormatBytes(memInfo.VMS)
		data.HWM = common.FormatBytes(memInfo.HWM)
		data.Data = common.FormatBytes(memInfo.Data)
		data.Stack = common.FormatBytes(memInfo.Stack)
		data.Locked = common.FormatBytes(memInfo.Locked)
		data.Swap = common.FormatBytes(memInfo.Swap)
		data.RssValue = memInfo.RSS
	} else {
		data.Rss = "--"
		data.Data = "--"
		data.VMS = "--"
		data.HWM = "--"
		data.Stack = "--"
		data.Locked = "--"
		data.Swap = "--"
		data.RssValue = 0
	}

	if envs, err := p.Environ(); err == nil {
		data.Envs = envs
	}

	if openFiles, err := p.OpenFiles(); err == nil {
		data.OpenFiles = openFiles
	}

	return data, nil
}
