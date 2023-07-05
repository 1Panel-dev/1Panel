package websocket

import (
	"encoding/json"
	"fmt"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
	"strings"
	"time"
)

type WsInput struct {
	Type string `json:"type"`
	DownloadProgress
	PsProcessConfig
	SSHSessionConfig
	NetConfig
}

type DownloadProgress struct {
	Keys []string `json:"keys"`
}

type PsProcessConfig struct {
	Pid      int32  `json:"pid"`
	Name     string `json:"name"`
	Username string `json:"username"`
}

type SSHSessionConfig struct {
	LoginUser string `json:"loginUser"`
	LoginIP   string `json:"loginIP"`
}

type NetConfig struct {
	Port        uint32 `json:"port"`
	ProcessName string `json:"processName"`
	ProcessID   int32  `json:"processID"`
}

type PsProcessData struct {
	PID            int32  `json:"PID"`
	Name           string `json:"name"`
	PPID           int32  `json:"PPID"`
	Username       string `json:"username"`
	Status         string `json:"status"`
	StartTime      string `json:"startTime"`
	NumThreads     int32  `json:"numThreads"`
	NumConnections int    `json:"numConnections"`
	CpuPercent     string `json:"cpuPercent"`

	DiskRead  string `json:"diskRead"`
	DiskWrite string `json:"diskWrite"`
	CmdLine   string `json:"cmdLine"`

	Rss    string `json:"rss"`
	VMS    string `json:"vms"`
	HWM    string `json:"hwm"`
	Data   string `json:"data"`
	Stack  string `json:"stack"`
	Locked string `json:"locked"`
	Swap   string `json:"swap"`

	CpuValue float64 `json:"cpuValue"`
	RssValue uint64  `json:"rssValue"`

	Envs []string `json:"envs"`

	OpenFiles []process.OpenFilesStat `json:"openFiles"`
	Connects  []processConnect        `json:"connects"`
}

type processConnect struct {
	Type   string   `json:"type"`
	Status string   `json:"status"`
	Laddr  net.Addr `json:"localaddr"`
	Raddr  net.Addr `json:"remoteaddr"`
	PID    int32    `json:"PID"`
	Name   string   `json:"name"`
}

type sshSession struct {
	Username  string `json:"username"`
	PID       int32  `json:"PID"`
	Terminal  string `json:"terminal"`
	Host      string `json:"host"`
	LoginTime string `json:"loginTime"`
}

func ProcessData(c *Client, inputMsg []byte) {
	wsInput := &WsInput{}
	err := json.Unmarshal(inputMsg, wsInput)
	if err != nil {
		global.LOG.Errorf("unmarshal wsInput error,err %s", err.Error())
		return
	}
	switch wsInput.Type {
	case "wget":
		res, err := getDownloadProcess(wsInput.DownloadProgress)
		if err != nil {
			return
		}
		c.Msg <- res
	case "ps":
		res, err := getProcessData(wsInput.PsProcessConfig)
		if err != nil {
			return
		}
		c.Msg <- res
	case "ssh":
		res, err := getSSHSessions(wsInput.SSHSessionConfig)
		if err != nil {
			return
		}
		c.Msg <- res
	case "net":
		res, err := getNetConnections(wsInput.NetConfig)
		if err != nil {
			return
		}
		c.Msg <- res
	}

}

func getDownloadProcess(progress DownloadProgress) (res []byte, err error) {
	var result []files.Process
	for _, k := range progress.Keys {
		value, err := global.CACHE.Get(k)
		if err != nil {
			global.LOG.Errorf("get cache error,err %s", err.Error())
			return nil, err
		}
		downloadProcess := &files.Process{}
		_ = json.Unmarshal(value, downloadProcess)
		result = append(result, *downloadProcess)
	}
	res, err = json.Marshal(result)
	return
}

const (
	b  = uint64(1)
	kb = 1024 * b
	mb = 1024 * kb
	gb = 1024 * mb
)

func formatBytes(bytes uint64) string {
	switch {
	case bytes < kb:
		return fmt.Sprintf("%dB", bytes)
	case bytes < mb:
		return fmt.Sprintf("%.2fKB", float64(bytes)/float64(kb))
	case bytes < gb:
		return fmt.Sprintf("%.2fMB", float64(bytes)/float64(mb))
	default:
		return fmt.Sprintf("%.2fGB", float64(bytes)/float64(gb))
	}
}

func getProcessData(processConfig PsProcessConfig) (res []byte, err error) {
	var (
		result    []PsProcessData
		processes []*process.Process
	)
	processes, err = process.Processes()
	if err != nil {
		return
	}
	for _, proc := range processes {
		procData := PsProcessData{
			PID: proc.Pid,
		}
		if processConfig.Pid > 0 && processConfig.Pid != proc.Pid {
			continue
		}
		if procName, err := proc.Name(); err == nil {
			procData.Name = procName
		} else {
			procData.Name = "<UNKNOWN>"
		}
		if processConfig.Name != "" && !strings.Contains(procData.Name, processConfig.Name) {
			continue
		}
		if username, err := proc.Username(); err == nil {
			procData.Username = username
		}
		if processConfig.Username != "" && !strings.Contains(procData.Username, processConfig.Username) {
			continue
		}
		procData.PPID, _ = proc.Ppid()
		statusArray, _ := proc.Status()
		if len(statusArray) > 0 {
			procData.Status = strings.Join(statusArray, ",")
		}
		createTime, procErr := proc.CreateTime()
		if procErr == nil {
			t := time.Unix(createTime/1000, 0)
			procData.StartTime = t.Format("2006-1-2 15:04:05")
		}
		procData.NumThreads, _ = proc.NumThreads()
		connections, procErr := proc.Connections()
		if procErr == nil {
			procData.NumConnections = len(connections)
			for _, conn := range connections {
				if conn.Laddr.IP != "" || conn.Raddr.IP != "" {
					procData.Connects = append(procData.Connects, processConnect{
						Status: conn.Status,
						Laddr:  conn.Laddr,
						Raddr:  conn.Raddr,
					})
				}
			}
		}
		procData.CpuValue, _ = proc.CPUPercent()
		procData.CpuPercent = fmt.Sprintf("%.2f", procData.CpuValue) + "%"
		menInfo, procErr := proc.MemoryInfo()
		if procErr == nil {
			procData.Rss = formatBytes(menInfo.RSS)
			procData.RssValue = menInfo.RSS
			procData.Data = formatBytes(menInfo.Data)
			procData.VMS = formatBytes(menInfo.VMS)
			procData.HWM = formatBytes(menInfo.HWM)
			procData.Stack = formatBytes(menInfo.Stack)
			procData.Locked = formatBytes(menInfo.Locked)
			procData.Swap = formatBytes(menInfo.Swap)
		} else {
			procData.Rss = "--"
			procData.Data = "--"
			procData.VMS = "--"
			procData.HWM = "--"
			procData.Stack = "--"
			procData.Locked = "--"
			procData.Swap = "--"

			procData.RssValue = 0
		}
		ioStat, procErr := proc.IOCounters()
		if procErr == nil {
			procData.DiskWrite = formatBytes(ioStat.WriteBytes)
			procData.DiskRead = formatBytes(ioStat.ReadBytes)
		} else {
			procData.DiskWrite = "--"
			procData.DiskRead = "--"
		}
		procData.CmdLine, _ = proc.Cmdline()
		procData.OpenFiles, _ = proc.OpenFiles()
		procData.Envs, _ = proc.Environ()

		result = append(result, procData)
	}
	res, err = json.Marshal(result)
	return
}

func getSSHSessions(config SSHSessionConfig) (res []byte, err error) {
	var (
		result    []sshSession
		users     []host.UserStat
		processes []*process.Process
	)
	processes, err = process.Processes()
	if err != nil {
		return
	}
	users, err = host.Users()
	if err != nil {
		return
	}
	for _, proc := range processes {
		name, _ := proc.Name()
		if name != "sshd" || proc.Pid == 0 {
			continue
		}
		connections, _ := proc.Connections()
		for _, conn := range connections {
			for _, user := range users {
				if user.Host == "" {
					continue
				}
				if conn.Raddr.IP == user.Host {
					if config.LoginUser != "" && !strings.Contains(user.User, config.LoginUser) {
						continue
					}
					if config.LoginIP != "" && !strings.Contains(user.Host, config.LoginIP) {
						continue
					}
					if terminal, err := proc.Cmdline(); err == nil {
						if strings.Contains(terminal, user.Terminal) {
							session := sshSession{
								Username: user.User,
								Host:     user.Host,
								Terminal: user.Terminal,
								PID:      proc.Pid,
							}
							t := time.Unix(int64(user.Started), 0)
							session.LoginTime = t.Format("2006-1-2 15:04:05")
							result = append(result, session)
						}
					}
				}
			}
		}
	}
	res, err = json.Marshal(result)
	return
}

var netTypes = [...]string{"tcp", "udp"}

func getNetConnections(config NetConfig) (res []byte, err error) {
	var (
		result []processConnect
		proc   *process.Process
	)
	for _, netType := range netTypes {
		connections, _ := net.Connections(netType)
		if err == nil {
			for _, conn := range connections {
				if config.ProcessID > 0 && config.ProcessID != conn.Pid {
					continue
				}
				proc, err = process.NewProcess(conn.Pid)
				if err == nil {
					name, _ := proc.Name()
					if name != "" && config.ProcessName != "" && !strings.Contains(name, config.ProcessName) {
						continue
					}
					if config.Port > 0 && config.Port != conn.Laddr.Port && config.Port != conn.Raddr.Port {
						continue
					}
					result = append(result, processConnect{
						Type:   netType,
						Status: conn.Status,
						Laddr:  conn.Laddr,
						Raddr:  conn.Raddr,
						PID:    conn.Pid,
						Name:   name,
					})
				}

			}
		}
	}
	res, err = json.Marshal(result)
	return
}
