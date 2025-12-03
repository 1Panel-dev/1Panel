package psutil

import (
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
		d := detectLinuxDistro()
		if strings.Contains(d, "(") && strings.Contains(d, ")") {
			d = d[:strings.LastIndex(d, "(")]
		}
		h.cachedDistro = strings.TrimSpace(d)
	}
	return h.cachedDistro
}

func detectLinuxDistro() string {
	distroFiles := []string{
		"/etc/os-release",
		"/usr/lib/os-release",
		"/etc/lsb-release",
		"/etc/redhat-release",
		"/etc/debian_version",
		"/etc/issue",
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
			switch targetFile {
			case "/etc/os-release", "/usr/lib/os-release":
				if v := findKeyValues(content, "PRETTY_NAME"); v["PRETTY_NAME"] != "" {
					return v["PRETTY_NAME"]
				}
				if v := findKeyValues(content, "NAME", "VERSION_ID"); v["NAME"] != "" && v["VERSION_ID"] != "" {
					return v["NAME"] + " " + v["VERSION_ID"]
				}
			case "/etc/lsb-release":
				if v := findKeyValues(content, "DISTRIB_DESCRIPTION"); v["DISTRIB_DESCRIPTION"] != "" {
					return v["DISTRIB_DESCRIPTION"]
				}
				if v := findKeyValues(content, "DISTRIB_ID", "DISTRIB_RELEASE"); v["DISTRIB_ID"] != "" && v["DISTRIB_RELEASE"] != "" {
					return v["DISTRIB_ID"] + " " + v["DISTRIB_RELEASE"]
				}
			case "/etc/redhat-release", "/etc/issue":
				return strings.TrimSpace(content)
			case "/etc/debian_version":
				return "Debian " + strings.TrimSpace(content)
			}
		}
	}

	// gopsutil fallback
	if osInfo, err := host.Info(); err == nil {
		return osInfo.OS
	}

	return "Unknown Linux"
}

func findKeyValues(data string, keys ...string) map[string]string {
	result := make(map[string]string, len(keys))
	found := 0
	for _, line := range strings.Split(data, "\n") {
		idx := strings.Index(line, "=")
		if idx == -1 {
			continue
		}
		key := line[:idx]
		for _, k := range keys {
			if key == k {
				result[k] = strings.Trim(line[idx+1:], "\"")
				found++
				if found == len(keys) {
					return result
				}
				break
			}
		}
	}
	return result
}
