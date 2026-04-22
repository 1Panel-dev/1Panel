package psutil

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/host"
)

const hostRefreshInterval = 4 * time.Hour

type HostInfoState struct {
	mu             sync.RWMutex
	lastSampleTime time.Time

	cachedInfo   *host.InfoStat
	cachedDistro string
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

func (h *HostInfoState) GetDistro() string {
	if h.cachedDistro == "" {
		h.cachedDistro = detectLinuxDistro()
	}
	return h.cachedDistro
}

func detectLinuxDistro() string {
	distroFiles := []string{
		"/etc/os-release",
		"/usr/lib/os-release",
	}

	var targetFile string
	for _, f := range distroFiles {
		if _, err := os.Stat(f); err == nil {
			targetFile = f
			break
		}
	}

	if targetFile != "" {
		data, err := os.ReadFile(targetFile)
		if err == nil {
			content := string(data)
			for _, line := range strings.Split(content, "\n") {
				idx := strings.Index(line, "=")
				if idx == -1 {
					continue
				}
				key := line[:idx]
				if key == "PRETTY_NAME" {
					d := strings.Trim(line[idx+1:], "\"")
					if strings.Contains(d, "(") && strings.Contains(d, ")") {
						d = d[:strings.LastIndex(d, "(")]
					}
					return strings.TrimSpace(d)
				}
			}
		}
	}

	if osInfo, err := host.Info(); err == nil {
		return fmt.Sprintf("%s %s", osInfo.Platform, osInfo.PlatformVersion)
	}

	return "Linux"
}
