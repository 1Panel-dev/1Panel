package psutil

import (
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/disk"
)

const (
	diskUsageCacheInterval     = 30 * time.Second
	diskPartitionCacheInterval = 10 * time.Minute
)

type DiskUsageEntry struct {
	lastSampleTime time.Time
	cachedUsage    *disk.UsageStat
}

type DiskState struct {
	usageMu    sync.RWMutex
	usageCache map[string]*DiskUsageEntry

	partitionMu       sync.RWMutex
	lastPartitionTime time.Time
	cachedPartitions  []disk.PartitionStat
}

func (d *DiskState) GetUsage(path string, forceRefresh bool) (*disk.UsageStat, error) {
	d.usageMu.RLock()
	if entry, ok := d.usageCache[path]; ok {
		if time.Since(entry.lastSampleTime) < diskUsageCacheInterval && !forceRefresh {
			defer d.usageMu.RUnlock()
			return entry.cachedUsage, nil
		}
	}
	d.usageMu.RUnlock()

	usage, err := disk.Usage(path)
	if err != nil {
		return nil, err
	}

	d.usageMu.Lock()
	if d.usageCache == nil {
		d.usageCache = make(map[string]*DiskUsageEntry)
	}
	d.usageCache[path] = &DiskUsageEntry{
		lastSampleTime: time.Now(),
		cachedUsage:    usage,
	}
	d.usageMu.Unlock()

	return usage, nil
}

func (d *DiskState) GetPartitions(all bool, forceRefresh bool) ([]disk.PartitionStat, error) {
	d.partitionMu.RLock()
	if d.cachedPartitions != nil && time.Since(d.lastPartitionTime) < diskPartitionCacheInterval && !forceRefresh {
		defer d.partitionMu.RUnlock()
		return d.cachedPartitions, nil
	}
	d.partitionMu.RUnlock()

	partitions, err := disk.Partitions(all)
	if err != nil {
		return nil, err
	}

	d.partitionMu.Lock()
	d.cachedPartitions = partitions
	d.lastPartitionTime = time.Now()
	d.partitionMu.Unlock()

	return partitions, nil
}

func (d *DiskState) ClearUsageCache(path string) {
	d.usageMu.Lock()
	delete(d.usageCache, path)
	d.usageMu.Unlock()
}

func (d *DiskState) ClearAllCache() {
	d.usageMu.Lock()
	d.usageCache = make(map[string]*DiskUsageEntry)
	d.usageMu.Unlock()

	d.partitionMu.Lock()
	d.cachedPartitions = nil
	d.partitionMu.Unlock()
}
