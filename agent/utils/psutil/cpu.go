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

type CPUUsageState struct {
	mu             sync.Mutex
	lastTotalStat  *CPUStat
	lastPerCPUStat []CPUStat
	lastSampleTime time.Time

	cachedTotalUsage float64
	cachedPerCore    []float64
}

func readCPUStat() (CPUStat, error) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return CPUStat{}, err
	}

	fields := strings.Fields(strings.Split(string(data), "\n")[0])[1:]
	nums := make([]uint64, len(fields))

	for i, f := range fields {
		v, _ := strconv.ParseUint(f, 10, 64)
		nums[i] = v
	}

	idle := nums[3] + nums[4]
	var total uint64
	for _, v := range nums {
		total += v
	}

	return CPUStat{Idle: idle, Total: total}, nil
}

func (c *CPUUsageState) readPerCPUStat() ([]CPUStat, error) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	stats := c.lastPerCPUStat[:0]

	for _, l := range lines[1:] {
		if !strings.HasPrefix(l, "cpu") {
			continue
		}
		if len(l) < 4 || l[3] < '0' || l[3] > '9' {
			continue
		}

		fields := strings.Fields(l)[1:]
		nums := make([]uint64, len(fields))
		for i, f := range fields {
			v, _ := strconv.ParseUint(f, 10, 64)
			nums[i] = v
		}

		idle := nums[3] + nums[4]
		var total uint64
		for _, v := range nums {
			total += v
		}

		stats = append(stats, CPUStat{Idle: idle, Total: total})
	}

	return stats, nil
}

func readPerCPUStatCopy() []CPUStat {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return nil
	}

	lines := strings.Split(string(data), "\n")
	var stats []CPUStat

	for _, l := range lines[1:] {
		if !strings.HasPrefix(l, "cpu") {
			continue
		}
		if len(l) < 4 || l[3] < '0' || l[3] > '9' {
			continue
		}

		fields := strings.Fields(l)[1:]
		nums := make([]uint64, len(fields))
		for i, f := range fields {
			v, _ := strconv.ParseUint(f, 10, 64)
			nums[i] = v
		}

		idle := nums[3] + nums[4]
		var total uint64
		for _, v := range nums {
			total += v
		}

		stats = append(stats, CPUStat{Idle: idle, Total: total})
	}

	return stats
}

func calcCPUPercent(prev, cur CPUStat) float64 {
	deltaIdle := float64(cur.Idle - prev.Idle)
	deltaTotal := float64(cur.Total - prev.Total)
	if deltaTotal <= 0 {
		return 0
	}
	return (1 - deltaIdle/deltaTotal) * 100
}

func (c *CPUUsageState) GetCPUUsage() (float64, []float64) {
	c.mu.Lock()

	now := time.Now()

	if !c.lastSampleTime.IsZero() && now.Sub(c.lastSampleTime) < fastInterval {
		result := c.cachedTotalUsage
		perCore := c.cachedPerCore
		c.mu.Unlock()
		return result, perCore
	}

	needReset := c.lastSampleTime.IsZero() || now.Sub(c.lastSampleTime) >= resetInterval
	c.mu.Unlock()

	if needReset {
		firstTotal, _ := readCPUStat()
		firstPer := readPerCPUStatCopy()
		time.Sleep(100 * time.Millisecond)
		secondTotal, _ := readCPUStat()
		secondPer := readPerCPUStatCopy()

		totalUsage := calcCPUPercent(firstTotal, secondTotal)

		perCore := make([]float64, len(secondPer))
		for i := range secondPer {
			perCore[i] = calcCPUPercent(firstPer[i], secondPer[i])
		}

		c.mu.Lock()
		c.cachedTotalUsage = totalUsage
		c.cachedPerCore = perCore
		c.lastTotalStat = &secondTotal
		c.lastPerCPUStat = secondPer
		c.lastSampleTime = time.Now()
		c.mu.Unlock()

		return totalUsage, perCore
	}

	curTotal, _ := readCPUStat()
	curPer := readPerCPUStatCopy()

	c.mu.Lock()
	defer c.mu.Unlock()

	totalUsage := calcCPUPercent(*c.lastTotalStat, curTotal)

	if len(c.cachedPerCore) != len(curPer) {
		c.cachedPerCore = make([]float64, len(curPer))
	}
	for i := range curPer {
		c.cachedPerCore[i] = calcCPUPercent(c.lastPerCPUStat[i], curPer[i])
	}

	c.cachedTotalUsage = totalUsage
	c.lastTotalStat = &curTotal
	c.lastPerCPUStat = curPer
	c.lastSampleTime = time.Now()

	return totalUsage, c.cachedPerCore
}

func (c *CPUUsageState) NumCPU() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return len(c.cachedPerCore)
}

type CPUInfoState struct {
	mu               sync.RWMutex
	initialized      bool
	cachedInfo       []cpu.InfoStat
	cachedPhysCores  int
	cachedLogicCores int
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
