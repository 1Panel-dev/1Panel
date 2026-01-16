package service

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/websocket"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

type ProcessService struct{}

type IProcessService interface {
	StopProcess(req request.ProcessReq) error
	GetProcessInfoByPID(pid int32) (*websocket.PsProcessData, error)
	GetListeningProcess(c context.Context) ([]ListeningProcess, error)
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

type ListeningProcess struct {
	PID      int32
	Port     map[uint32]struct{}
	Protocol uint32
	Name     string
}

func (ps *ProcessService) GetListeningProcess(c context.Context) ([]ListeningProcess, error) {
	conn, err := net.ConnectionsMaxWithContext(c, "inet", 32768)
	if err != nil {
		return nil, err
	}
	procCache := make(map[int32]ListeningProcess, 64)

	for _, conn := range conn {
		if conn.Pid == 0 {
			continue
		}

		if (conn.Status == "LISTEN" && conn.Type == syscall.SOCK_STREAM) || (conn.Type == syscall.SOCK_DGRAM && conn.Raddr.Port == 0) {
			if _, exists := procCache[conn.Pid]; !exists {
				proc, err := process.NewProcess(conn.Pid)
				if err != nil {
					continue
				}
				procData := ListeningProcess{
					PID: conn.Pid,
				}
				procData.Name, _ = proc.Name()
				procData.Port = make(map[uint32]struct{})
				procData.Port[conn.Laddr.Port] = struct{}{}
				procData.Protocol = conn.Type
				procCache[conn.Pid] = procData
			} else {
				p := procCache[conn.Pid]
				p.Port[conn.Laddr.Port] = struct{}{}
				procCache[conn.Pid] = p
			}
		}
	}

	procs := make([]ListeningProcess, 0, len(procCache))
	for _, proc := range procCache {
		procs = append(procs, proc)
	}

	return procs, nil
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

	if memDetail, err := getMemoryDetail(p.Pid); err == nil {
		data.Rss = common.FormatBytes(memDetail.RSS)
		data.VMS = common.FormatBytes(memDetail.VMS)
		data.HWM = common.FormatBytes(memDetail.HWM)
		data.Data = common.FormatBytes(memDetail.Data)
		data.Stack = common.FormatBytes(memDetail.Stack)
		data.Locked = common.FormatBytes(memDetail.Locked)
		data.Swap = common.FormatBytes(memDetail.Swap)
		data.Dirty = common.FormatBytes(memDetail.Dirty)
		data.RssValue = memDetail.RSS
		data.PSS = common.FormatBytes(memDetail.PSS)
		data.USS = common.FormatBytes(memDetail.USS)
		data.Shared = common.FormatBytes(memDetail.Shared)
		data.Text = common.FormatBytes(memDetail.Text)
	}

	if envs, err := p.Environ(); err == nil {
		data.Envs = envs
	}

	if openFiles, err := p.OpenFiles(); err == nil {
		data.OpenFiles = openFiles
	}

	return data, nil
}

type MemoryDetail struct {
	RSS    uint64
	VMS    uint64
	HWM    uint64
	Data   uint64
	Stack  uint64
	Locked uint64
	Swap   uint64

	PSS    uint64
	USS    uint64
	Shared uint64
	Text   uint64
	Dirty  uint64
}

func getMemoryDetail(pid int32) (*MemoryDetail, error) {
	mem := &MemoryDetail{}

	if err := readStatus(pid, mem); err != nil {
		return nil, err
	}

	if err := readSmapsRollup(pid, mem); err != nil {
		if err := readSmaps(pid, mem); err != nil {
			return nil, err
		}
	}
	return mem, nil
}

func readStatus(pid int32, mem *MemoryDetail) error {
	filePath := fmt.Sprintf("/proc/%d/status", pid)
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		key := strings.TrimSuffix(fields[0], ":")
		value, _ := strconv.ParseUint(fields[1], 10, 64)
		value *= 1024

		switch key {
		case "VmRSS":
			mem.RSS = value
		case "VmSize":
			mem.VMS = value
		case "VmData":
			mem.Data = value
		case "VmSwap":
			mem.Swap = value
		case "VmExe":
			mem.Text = value
		case "RssShmem":
			mem.Shared = value
		case "VmHWM":
			mem.HWM = value
		case "VmStk":
			mem.Stack = value
		case "VmLck":
			mem.Locked = value
		}
	}

	return scanner.Err()
}

func readSmapsRollup(pid int32, mem *MemoryDetail) error {
	filePath := fmt.Sprintf("/proc/%d/smaps_rollup", pid)
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		key := strings.TrimSuffix(fields[0], ":")
		value, _ := strconv.ParseUint(fields[1], 10, 64)
		value *= 1024

		switch key {
		case "Pss":
			mem.PSS = value
		case "Private_Clean", "Private_Dirty":
			mem.USS += value
		case "Shared_Clean", "Shared_Dirty":
			if mem.Shared == 0 {
				mem.Shared = value
			}
		}
	}

	return scanner.Err()
}

func readSmaps(pid int32, mem *MemoryDetail) error {
	filePath := fmt.Sprintf("/proc/%d/smaps", pid)
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		key := strings.TrimSuffix(fields[0], ":")
		value, _ := strconv.ParseUint(fields[1], 10, 64)
		value *= 1024

		switch key {
		case "Pss":
			mem.PSS += value
		case "Private_Clean", "Private_Dirty":
			mem.USS += value
		case "Shared_Clean", "Shared_Dirty":
			if mem.Shared == 0 {
				mem.Shared += value
			}
		}
	}

	return scanner.Err()
}
