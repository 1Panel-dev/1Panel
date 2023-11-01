package ps

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/process"
)

func TestPs(t *testing.T) {
	processes, err := process.Processes()
	if err != nil {
		panic(err)
	}
	for _, pro := range processes {
		var (
			name           string
			parentID       int32
			userName       string
			status         string
			startTime      string
			numThreads     int32
			numConnections int
			cpuPercent     float64
			//mem            string
			rss     string
			ioRead  string
			ioWrite string
		)
		name, _ = pro.Name()
		parentID, _ = pro.Ppid()
		userName, _ = pro.Username()
		array, err := pro.Status()
		if err == nil {
			status = array[0]
		}
		createTime, err := pro.CreateTime()
		if err == nil {
			t := time.Unix(createTime/1000, 0)
			startTime = t.Format("2006-1-2 15:04:05")
		}
		numThreads, _ = pro.NumThreads()
		connections, err := pro.Connections()
		if err == nil && len(connections) > 0 {
			numConnections = len(connections)
		}
		cpuPercent, _ = pro.CPUPercent()
		menInfo, err := pro.MemoryInfo()
		if err == nil {
			rssF := float64(menInfo.RSS) / 1048576
			rss = fmt.Sprintf("%.2f", rssF)
		}
		ioStat, err := pro.IOCounters()
		if err == nil {
			ioWrite = strconv.FormatUint(ioStat.WriteBytes, 10)
			ioRead = strconv.FormatUint(ioStat.ReadBytes, 10)
		}

		cmdLine, err := pro.Cmdline()
		if err == nil {
			fmt.Println(cmdLine)
		}
		ss, err := pro.Terminal()
		if err == nil {
			fmt.Println(ss)
		}

		fmt.Printf("Name: %s PId: %v ParentID: %v Username: %v status:%s startTime: %s numThreads: %v numConnections:%v cpuPercent:%v rss:%s MB IORead: %s IOWrite: %s \n",
			name, pro.Pid, parentID, userName, status, startTime, numThreads, numConnections, cpuPercent, rss, ioRead, ioWrite)
	}
	users, err := host.Users()
	if err == nil {
		fmt.Println(users)
	}

}
