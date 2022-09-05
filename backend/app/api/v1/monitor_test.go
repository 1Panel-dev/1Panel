package v1

import (
	"fmt"
	"testing"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
)

func TestMonito(t *testing.T) {
	totalPercent, _ := cpu.Percent(3*time.Second, false) // 总 cpu 使用
	perPercents, _ := cpu.Percent(3*time.Second, true)   // 各 cpu 使用
	fmt.Println("================totalPercent============", totalPercent)
	fmt.Println("================perPercents=============", perPercents)

	info, _ := load.Avg()
	info2, _ := load.Misc()
	fmt.Printf("load: \n loadxx: %v   load1: %v, load5: %v, load15: %v \n\n", info2, info.Load1, info.Load5, info.Load15)

	memory, _ := mem.VirtualMemory()
	fmt.Printf("memory: \n   memory used: %v, use persent: %v \n\n", memory.Used, memory.UsedPercent)

	diskPart, _ := disk.Partitions(true)
	// fmt.Println("================disk=============", diskPart)
	for _, part := range diskPart {
		diskInfo, _ := disk.Usage(part.Mountpoint)
		fmt.Printf("inode,disk(%v): \n    inode persent: %v, disk used: %v, persent: %v \n", part.Mountpoint, diskInfo.InodesUsedPercent, diskInfo.Used, diskInfo.UsedPercent)
	}
	fmt.Println()

	ioStat, _ := disk.IOCounters()
	for _, v := range ioStat {
		fmt.Printf("io: \n    name: %v, readCount: %v, writeCount: %v \n\n", v.Name, v.ReadCount, v.WriteCount)
	}
	fmt.Println()

	net, _ := net.IOCounters(false)
	for _, v := range net {
		fmt.Printf("netio: \n    %v: send:%v recv:%v \n\n", v.Name, v.BytesSent, v.BytesRecv)
	}
	fmt.Println()
}
