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

type CPUStat struct {
	Idle  uint64
	Total uint64
}

type CPUDetailedStat struct {
	User      uint64
	Nice      uint64
	System    uint64
	Idle      uint64
	Iowait    uint64
	Irq       uint64
	Softirq   uint64
	Steal     uint64
	Guest     uint64
	GuestNice uint64
	Total     uint64
}

type CPUDetailedPercent struct {
	User    float64 `json:"user"`
	System  float64 `json:"system"`
	Nice    float64 `json:"nice"`
	Idle    float64 `json:"idle"`
	Iowait  float64 `json:"iowait"`
	Irq     float64 `json:"irq"`
	Softirq float64 `json:"softirq"`
	Steal   float64 `json:"steal"`
}

func (c *CPUDetailedPercent) GetCPUDetailedPercent() []float64 {
	return []float64{c.User, c.System, c.Nice, c.Idle, c.Iowait, c.Irq, c.Softirq, c.Steal}
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

func (c *CPUUsageState) GetCPUUsage() (float64, []float64, []float64) {
	c.mu.Lock()

	now := time.Now()

	if !c.lastSampleTime.IsZero() && now.Sub(c.lastSampleTime) < fastInterval {
		result := c.cachedTotalUsage
		perCore := c.cachedPerCore
		detailed := c.cachedDetailedPercent
		c.mu.Unlock()
		return result, perCore, detailed.GetCPUDetailedPercent()
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

		return totalUsage, perCore, detailedPercent.GetCPUDetailedPercent()
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

	return totalUsage, c.cachedPerCore, detailedPercent.GetCPUDetailedPercent()
}

func (c *CPUUsageState) NumCPU() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return len(c.cachedPerCore)
}

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

func readProcStat() ([]byte, error) {
	return os.ReadFile("/proc/stat")
}

func parseCPUFields(line string) []uint64 {
	fields := strings.Fields(line)
	if len(fields) <= 1 {
		return nil
	}
	fields = fields[1:]

	nums := make([]uint64, len(fields))
	for i, f := range fields {
		v, _ := strconv.ParseUint(f, 10, 64)
		nums[i] = v
	}
	return nums
}

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

func readAllCPUStat() (CPUStat, CPUDetailedStat, []CPUStat) {
	data, err := readProcStat()
	if err != nil {
		return CPUStat{}, CPUDetailedStat{}, nil
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 {
		return CPUStat{}, CPUDetailedStat{}, nil
	}

	firstLine := lines[0]
	nums := parseCPUFields(firstLine)

	idle, total := calcIdleAndTotal(nums)
	cpuStat := CPUStat{Idle: idle, Total: total}

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
	detailedStat.Total = detailedStat.User + detailedStat.Nice + detailedStat.System +
		detailedStat.Idle + detailedStat.Iowait + detailedStat.Irq + detailedStat.Softirq + detailedStat.Steal

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

func calcCPUPercent(prev, cur CPUStat) float64 {
	deltaIdle := float64(cur.Idle - prev.Idle)
	deltaTotal := float64(cur.Total - prev.Total)
	if deltaTotal <= 0 {
		return 0
	}
	return (1 - deltaIdle/deltaTotal) * 100
}

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
