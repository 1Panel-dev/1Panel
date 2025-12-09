package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

const defaultTimeout = 10 * time.Second

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
	Dirty  string `json:"dirty"`
	PSS    string `json:"pss"`
	USS    string `json:"uss"`
	Shared string `json:"shared"`
	Text   string `json:"text"`

	CpuValue float64 `json:"cpuValue"`
	RssValue uint64  `json:"rssValue"`

	Envs []string `json:"envs"`

	OpenFiles []process.OpenFilesStat `json:"openFiles"`
	Connects  []ProcessConnect        `json:"connects"`
}

type ProcessConnect struct {
	Type   string   `json:"type"`
	Status string   `json:"status"`
	Laddr  net.Addr `json:"localaddr"`
	Raddr  net.Addr `json:"remoteaddr"`
	PID    int32    `json:"PID"`
	Name   string   `json:"name"`
}

type ProcessConnects []ProcessConnect

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
		c.Send(res)
	case "ps":
		res, err := getProcessData(wsInput.PsProcessConfig)
		if err != nil {
			return
		}
		c.Send(res)
	case "ssh":
		res, err := getSSHSessions(wsInput.SSHSessionConfig)
		if err != nil {
			return
		}
		c.Send(res)
	case "net":
		res, err := getNetConnections(wsInput.NetConfig)
		if err != nil {
			return
		}
		c.Send(res)
	}

}

func getDownloadProcess(progress DownloadProgress) (res []byte, err error) {
	var result []files.Process
	for _, k := range progress.Keys {
		value := global.CACHE.Get(k)
		if value == "" {
			return nil, fmt.Errorf("get cache error,err value is nil")
		}
		downloadProcess := &files.Process{}
		_ = json.Unmarshal([]byte(value), downloadProcess)
		result = append(result, *downloadProcess)
		if downloadProcess.Percent == 100 {
			global.CACHE.Del(k)
		}
	}
	res, err = json.Marshal(result)
	return
}

func handleProcessData(proc *process.Process, processConfig *PsProcessConfig, pidConnections map[int32][]net.ConnectionStat) *PsProcessData {
	if processConfig.Pid > 0 && processConfig.Pid != proc.Pid {
		return nil
	}
	procData := PsProcessData{
		PID: proc.Pid,
	}
	if procName, err := proc.Name(); err == nil {
		procData.Name = procName
	} else {
		procData.Name = "<UNKNOWN>"
	}
	if processConfig.Name != "" && !strings.Contains(procData.Name, processConfig.Name) {
		return nil
	}
	if username, err := proc.Username(); err == nil {
		procData.Username = username
	}
	if processConfig.Username != "" && !strings.Contains(procData.Username, processConfig.Username) {
		return nil
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
	procData.CpuValue, _ = proc.CPUPercent()
	procData.CpuPercent = fmt.Sprintf("%.2f%%", procData.CpuValue)

	if memInfo, err := proc.MemoryInfo(); err == nil {
		procData.RssValue = memInfo.RSS
		procData.Rss = common.FormatBytes(memInfo.RSS)
	} else {
		procData.RssValue = 0
	}

	if connections, ok := pidConnections[proc.Pid]; ok {
		procData.NumConnections = len(connections)
	}

	return &procData
}

func getProcessData(processConfig PsProcessConfig) (res []byte, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	processes, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return
	}

	connections, err := net.ConnectionsMaxWithContext(ctx, "all", 32768)
	if err != nil {
		return
	}

	pidConnections := make(map[int32][]net.ConnectionStat, len(processes))
	for _, conn := range connections {
		if conn.Pid == 0 {
			continue
		}
		pidConnections[conn.Pid] = append(pidConnections[conn.Pid], conn)
	}

	result := make([]PsProcessData, 0, len(processes))

	for _, proc := range processes {
		procData := handleProcessData(proc, &processConfig, pidConnections)
		if procData != nil {
			result = append(result, *procData)
		}
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
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	users, err = host.UsersWithContext(ctx)
	if err != nil {
		res, err = json.Marshal(result)
		return
	}

	usersByHost := make(map[string][]host.UserStat)
	for _, user := range users {
		if user.Host == "" {
			continue
		}
		if config.LoginUser != "" && !strings.Contains(user.User, config.LoginUser) {
			continue
		}
		if config.LoginIP != "" && !strings.Contains(user.Host, config.LoginIP) {
			continue
		}
		usersByHost[user.Host] = append(usersByHost[user.Host], user)
	}

	if len(usersByHost) == 0 {
		res, err = json.Marshal(result)
		return
	}

	processes, err = process.ProcessesWithContext(ctx)
	if err != nil {
		res, err = json.Marshal(result)
		return
	}

	for _, proc := range processes {
		name, _ := proc.Name()
		if name != "sshd" || proc.Pid == 0 {
			continue
		}
		connections, _ := proc.Connections()
		if len(connections) == 0 {
			continue
		}

		cmdline, cmdErr := proc.Cmdline()
		if cmdErr != nil {
			continue
		}

		for _, conn := range connections {
			matchedUsers, exists := usersByHost[conn.Raddr.IP]
			if !exists {
				continue
			}

			for _, user := range matchedUsers {
				if strings.Contains(cmdline, user.Terminal) {
					t := time.Unix(int64(user.Started), 0)
					result = append(result, sshSession{
						Username:  user.User,
						Host:      user.Host,
						Terminal:  user.Terminal,
						PID:       proc.Pid,
						LoginTime: t.Format("2006-1-2 15:04:05"),
					})
				}
			}
		}
	}
	res, err = json.Marshal(result)
	return
}

func getNetConnections(config NetConfig) (res []byte, err error) {
	result := make([]ProcessConnect, 0, 1024)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	connections, err := net.ConnectionsMaxWithContext(ctx, "all", 32768)
	if err != nil {
		res, _ = json.Marshal(result)
		return
	}

	pidConnectionsMap := make(map[int32][]net.ConnectionStat, 256)
	pidNameMap := make(map[int32]string, 256)

	for _, conn := range connections {
		if conn.Family != 2 && conn.Family != 10 {
			continue
		}

		if conn.Pid == 0 {
			continue
		}

		if config.ProcessID > 0 && conn.Pid != config.ProcessID {
			continue
		}

		if config.Port > 0 && conn.Laddr.Port != config.Port && conn.Raddr.Port != config.Port {
			continue
		}

		if _, exists := pidNameMap[conn.Pid]; !exists {
			pName, _ := getProcessNameWithContext(ctx, conn.Pid)
			if pName == "" {
				pName = "<UNKNOWN>"
			}
			pidNameMap[conn.Pid] = pName
		}

		pidConnectionsMap[conn.Pid] = append(pidConnectionsMap[conn.Pid], conn)
	}

	for pid, connections := range pidConnectionsMap {
		pName := pidNameMap[pid]
		if config.ProcessName != "" && !strings.Contains(pName, config.ProcessName) {
			continue
		}
		for _, conn := range connections {
			result = append(result, ProcessConnect{
				Type:   getConnectionType(conn.Type, conn.Family),
				Status: conn.Status,
				Laddr:  conn.Laddr,
				Raddr:  conn.Raddr,
				PID:    conn.Pid,
				Name:   pName,
			})
		}
	}

	res, err = json.Marshal(result)
	return
}

func getProcessNameWithContext(ctx context.Context, pid int32) (string, error) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", pid))
	if err == nil && len(data) > 0 {
		return strings.TrimSpace(string(data)), nil
	}
	p, err := process.NewProcessWithContext(ctx, pid)
	if err != nil {
		return "", err
	}
	return p.Name()
}

func getConnectionType(connType uint32, family uint32) string {
	switch {
	case connType == 1 && family == 2:
		return "tcp"
	case connType == 1 && family == 10:
		return "tcp6"
	case connType == 2 && family == 2:
		return "udp"
	case connType == 2 && family == 10:
		return "udp6"
	default:
		return "unknown"
	}
}
