package psutil

import (
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/host"
)

const hostRefreshInterval = 4 * time.Hour

type HostInfoState struct {
	mu             sync.RWMutex
	lastSampleTime time.Time

	cachedInfo *host.InfoStat
}

func (h *HostInfoState) GetHostInfo(forceRefresh bool) (*host.InfoStat, error) {
	h.mu.RLock()
	if h.cachedInfo != nil && time.Since(h.lastSampleTime) < hostRefreshInterval && !forceRefresh {
		defer h.mu.RUnlock()
		return h.cachedInfo, nil
	}
	h.mu.RUnlock()

	hostInfo, err := host.Info()
	if err != nil {
		return nil, err
	}

	h.mu.Lock()
	h.cachedInfo = hostInfo
	h.lastSampleTime = time.Now()
	h.mu.Unlock()

	return hostInfo, nil
}
