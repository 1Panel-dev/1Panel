package psutil

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
)

const (
	resetInterval = 1 * time.Minute
	fastInterval  = 3 * time.Second
)

// ============================================================================
// 结构体定义
// ============================================================================

type CPUStat struct {
	Idle  uint64
	Total uint64
}

// CPUDetailedStat 存储 /proc/stat 中的详细 CPU 时间数据
// 字段顺序: user, nice, system, idle, iowait, irq, softirq, steal, guest, guest_nice
type CPUDetailedStat struct {
	User      uint64 // 用户态时间
	Nice      uint64 // 低优先级用户态时间
	System    uint64 // 内核态时间
	Idle      uint64 // 空闲时间
	Iowait    uint64 // I/O 等待时间
	Irq       uint64 // 硬中断时间
	Softirq   uint64 // 软中断时间
	Steal     uint64 // 虚拟化环境中被其他 OS 占用的时间
	Guest     uint64 // 运行虚拟 CPU 的时间
	GuestNice uint64 // 运行低优先级虚拟 CPU 的时间
	Total     uint64 // 总时间
}

// CPUDetailedPercent 存储类似 top 命令的 CPU 百分比信息
type CPUDetailedPercent struct {
	User    float64 `json:"user"`    // %us - 用户空间占用
	System  float64 `json:"system"`  // %sy - 内核空间占用
	Nice    float64 `json:"nice"`    // %ni - 改变过优先级的进程占用
	Idle    float64 `json:"idle"`    // %id - 空闲
	Iowait  float64 `json:"iowait"`  // %wa - I/O 等待
	Irq     float64 `json:"irq"`     // %hi - 硬中断
	Softirq float64 `json:"softirq"` // %si - 软中断
	Steal   float64 `json:"steal"`   // %st - 虚拟机偷取
}

type CPUUsageState struct {
	mu             sync.Mutex
	lastTotalStat  *CPUStat
	lastPerCPUStat []CPUStat
	lastDetailStat *CPUDetailedStat
	lastSampleTime time.Time

	cachedTotalUsage      float64
	cachedPerCore         []float64
	cachedDetailedPercent CPUDetailedPercent
}

type CPUInfoState struct {
	mu               sync.RWMutex
	initialized      bool
	cachedInfo       []cpu.InfoStat
	cachedPhysCores  int
	cachedLogicCores int
}

// ============================================================================
// CPUUsageState 公有方法
// ============================================================================

// GetCPUUsage 返回 CPU 使用率、每核使用率和详细百分比信息
// 返回: totalUsage, perCoreUsage, detailedPercent (类似 top 命令的 %us, %sy, %ni, %id, %wa, %hi, %si, %st)
func (c *CPUUsageState) GetCPUUsage() (float64, []float64, CPUDetailedPercent) {
	c.mu.Lock()

	now := time.Now()

	if !c.lastSampleTime.IsZero() && now.Sub(c.lastSampleTime) < fastInterval {
		result := c.cachedTotalUsage
		perCore := c.cachedPerCore
		detailed := c.cachedDetailedPercent
		c.mu.Unlock()
		return result, perCore, detailed
	}

	needReset := c.lastSampleTime.IsZero() || now.Sub(c.lastSampleTime) >= resetInterval
	c.mu.Unlock()

	if needReset {
		firstTotal, firstDetail, firstPer := readAllCPUStat()
		time.Sleep(100 * time.Millisecond)
		secondTotal, secondDetail, secondPer := readAllCPUStat()

		totalUsage := calcCPUPercent(firstTotal, secondTotal)
		detailedPercent := calcCPUDetailedPercent(firstDetail, secondDetail)

		perCore := make([]float64, len(secondPer))
		for i := range secondPer {
			perCore[i] = calcCPUPercent(firstPer[i], secondPer[i])
		}

		c.mu.Lock()
		c.cachedTotalUsage = totalUsage
		c.cachedPerCore = perCore
		c.cachedDetailedPercent = detailedPercent
		c.lastTotalStat = &secondTotal
		c.lastDetailStat = &secondDetail
		c.lastPerCPUStat = secondPer
		c.lastSampleTime = time.Now()
		c.mu.Unlock()

		return totalUsage, perCore, detailedPercent
	}

	curTotal, curDetail, curPer := readAllCPUStat()

	c.mu.Lock()
	defer c.mu.Unlock()

	totalUsage := calcCPUPercent(*c.lastTotalStat, curTotal)
	detailedPercent := calcCPUDetailedPercent(*c.lastDetailStat, curDetail)

	if len(c.cachedPerCore) != len(curPer) {
		c.cachedPerCore = make([]float64, len(curPer))
	}
	for i := range curPer {
		c.cachedPerCore[i] = calcCPUPercent(c.lastPerCPUStat[i], curPer[i])
	}

	c.cachedTotalUsage = totalUsage
	c.cachedPerCore = c.cachedPerCore
	c.cachedDetailedPercent = detailedPercent
	c.lastTotalStat = &curTotal
	c.lastDetailStat = &curDetail
	c.lastPerCPUStat = curPer
	c.lastSampleTime = time.Now()

	return totalUsage, c.cachedPerCore, detailedPercent
}

func (c *CPUUsageState) NumCPU() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return len(c.cachedPerCore)
}

// ============================================================================
// CPUInfoState 公有方法
// ============================================================================

func (c *CPUInfoState) GetCPUInfo(forceRefresh bool) ([]cpu.InfoStat, error) {
	c.mu.RLock()
	if c.initialized && c.cachedInfo != nil && !forceRefresh {
		defer c.mu.RUnlock()
		return c.cachedInfo, nil
	}
	c.mu.RUnlock()

	info, err := cpu.Info()
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.cachedInfo = info
	c.initialized = true
	c.mu.Unlock()

	return info, nil
}

func (c *CPUInfoState) GetPhysicalCores(forceRefresh bool) (int, error) {
	c.mu.RLock()
	if c.initialized && c.cachedPhysCores > 0 && !forceRefresh {
		defer c.mu.RUnlock()
		return c.cachedPhysCores, nil
	}
	c.mu.RUnlock()

	cores, err := cpu.Counts(false)
	if err != nil {
		return 0, err
	}

	c.mu.Lock()
	c.cachedPhysCores = cores
	c.initialized = true
	c.mu.Unlock()

	return cores, nil
}

func (c *CPUInfoState) GetLogicalCores(forceRefresh bool) (int, error) {
	c.mu.RLock()
	if c.initialized && c.cachedLogicCores > 0 && !forceRefresh {
		defer c.mu.RUnlock()
		return c.cachedLogicCores, nil
	}
	c.mu.RUnlock()

	cores, err := cpu.Counts(true)
	if err != nil {
		return 0, err
	}

	c.mu.Lock()
	c.cachedLogicCores = cores
	c.initialized = true
	c.mu.Unlock()

	return cores, nil
}

// ============================================================================
// 私有函数
// ============================================================================

// readProcStat 读取 /proc/stat 文件内容
func readProcStat() ([]byte, error) {
	return os.ReadFile("/proc/stat")
}

// parseCPUFields 解析 CPU 行的数值字段
func parseCPUFields(line string) []uint64 {
	fields := strings.Fields(line)
	if len(fields) <= 1 {
		return nil
	}
	fields = fields[1:] // 跳过 "cpu" 或 "cpuN" 前缀

	nums := make([]uint64, len(fields))
	for i, f := range fields {
		v, _ := strconv.ParseUint(f, 10, 64)
		nums[i] = v
	}
	return nums
}

// calcIdleAndTotal 计算空闲时间和总时间
func calcIdleAndTotal(nums []uint64) (idle, total uint64) {
	if len(nums) < 5 {
		return 0, 0
	}
	idle = nums[3] + nums[4]
	for _, v := range nums {
		total += v
	}
	return
}

// readAllCPUStat 一次性读取所有 CPU 统计数据，避免多次读取 /proc/stat
// 返回: 总CPU统计、详细CPU统计、每核CPU统计
func readAllCPUStat() (CPUStat, CPUDetailedStat, []CPUStat) {
	data, err := readProcStat()
	if err != nil {
		return CPUStat{}, CPUDetailedStat{}, nil
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 {
		return CPUStat{}, CPUDetailedStat{}, nil
	}

	// 解析第一行 (总 CPU)
	firstLine := lines[0]
	nums := parseCPUFields(firstLine)

	// CPUStat
	idle, total := calcIdleAndTotal(nums)
	cpuStat := CPUStat{Idle: idle, Total: total}

	// CPUDetailedStat - 确保至少有 10 个元素
	if len(nums) < 10 {
		padded := make([]uint64, 10)
		copy(padded, nums)
		nums = padded
	}
	detailedStat := CPUDetailedStat{
		User:      nums[0],
		Nice:      nums[1],
		System:    nums[2],
		Idle:      nums[3],
		Iowait:    nums[4],
		Irq:       nums[5],
		Softirq:   nums[6],
		Steal:     nums[7],
		Guest:     nums[8],
		GuestNice: nums[9],
	}
	// 计算总时间 (不包括 guest 和 guest_nice，因为它们已经包含在 user 和 nice 中)
	detailedStat.Total = detailedStat.User + detailedStat.Nice + detailedStat.System +
		detailedStat.Idle + detailedStat.Iowait + detailedStat.Irq + detailedStat.Softirq + detailedStat.Steal

	// 解析每核 CPU
	var perCPUStats []CPUStat
	for _, line := range lines[1:] {
		if !strings.HasPrefix(line, "cpu") {
			continue
		}
		if len(line) < 4 || line[3] < '0' || line[3] > '9' {
			continue
		}

		perNums := parseCPUFields(line)
		perIdle, perTotal := calcIdleAndTotal(perNums)
		perCPUStats = append(perCPUStats, CPUStat{Idle: perIdle, Total: perTotal})
	}

	return cpuStat, detailedStat, perCPUStats
}

// calcCPUPercent 计算两次采样之间的 CPU 使用率百分比
func calcCPUPercent(prev, cur CPUStat) float64 {
	deltaIdle := float64(cur.Idle - prev.Idle)
	deltaTotal := float64(cur.Total - prev.Total)
	if deltaTotal <= 0 {
		return 0
	}
	return (1 - deltaIdle/deltaTotal) * 100
}

// calcCPUDetailedPercent 根据两次采样计算各项 CPU 百分比
func calcCPUDetailedPercent(prev, cur CPUDetailedStat) CPUDetailedPercent {
	deltaTotal := float64(cur.Total - prev.Total)
	if deltaTotal <= 0 {
		return CPUDetailedPercent{Idle: 100}
	}

	return CPUDetailedPercent{
		User:    float64(cur.User-prev.User) / deltaTotal * 100,
		System:  float64(cur.System-prev.System) / deltaTotal * 100,
		Nice:    float64(cur.Nice-prev.Nice) / deltaTotal * 100,
		Idle:    float64(cur.Idle-prev.Idle) / deltaTotal * 100,
		Iowait:  float64(cur.Iowait-prev.Iowait) / deltaTotal * 100,
		Irq:     float64(cur.Irq-prev.Irq) / deltaTotal * 100,
		Softirq: float64(cur.Softirq-prev.Softirq) / deltaTotal * 100,
		Steal:   float64(cur.Steal-prev.Steal) / deltaTotal * 100,
	}
}
