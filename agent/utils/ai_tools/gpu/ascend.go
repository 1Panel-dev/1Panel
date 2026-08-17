package gpu

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/utils/ai_tools/gpu/common"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

var (
	ascendVersionPattern = regexp.MustCompile(`(?i)\bVersion:\s*([^\s|]+)`)
	ascendMemoryPattern  = regexp.MustCompile(`([0-9]+(?:\.[0-9]+)?)\s*/\s*([0-9]+(?:\.[0-9]+)?)`)
)

type AscendSMI struct{}

func (a AscendSMI) LoadGpuInfo() (*common.GpuInfo, error) {
	cmdMgr := cmd.NewCommandMgr(cmd.WithTimeout(5 * time.Second))
	itemData, err := cmdMgr.RunWithStdout("npu-smi", "info")
	if err != nil {
		return nil, fmt.Errorf("calling npu-smi failed, %v", err)
	}
	return parseAscendSMI(itemData), nil
}

func parseAscendSMI(data string) *common.GpuInfo {
	info := &common.GpuInfo{Type: "ascend"}
	if match := ascendVersionPattern.FindStringSubmatch(data); len(match) == 2 {
		info.DriverVersion = match[1]
	}

	processSection := false
	deviceIndexes := make(map[uint]int)
	var pending *common.GPU

	for _, line := range strings.Split(data, "\n") {
		if strings.Contains(line, "Process id") && strings.Contains(line, "Process memory") {
			processSection = true
			pending = nil
			continue
		}

		cells := ascendTableCells(line)
		if processSection {
			if len(cells) != 4 && len(cells) != 5 {
				continue
			}
			deviceFields := strings.Fields(cells[0])
			if len(deviceFields) == 0 {
				continue
			}
			deviceID, err := strconv.ParseUint(deviceFields[0], 10, 64)
			if err != nil {
				continue
			}
			index, ok := deviceIndexes[uint(deviceID)]
			if !ok {
				continue
			}
			processOffset := len(cells) - 3
			info.GPUs[index].Processes = append(info.GPUs[index].Processes, common.Process{
				Pid:         strings.TrimSpace(cells[processOffset]),
				Type:        "NPU",
				ProcessName: strings.TrimSpace(cells[processOffset+1]),
				UsedMemory:  ascendValueWithUnit(cells[processOffset+2], "MB"),
			})
			continue
		}

		if len(cells) != 3 {
			continue
		}
		if pending == nil {
			fields := strings.Fields(cells[0])
			if len(fields) < 2 {
				continue
			}
			deviceID, err := strconv.ParseUint(fields[0], 10, 64)
			if err != nil {
				continue
			}
			metrics := strings.Fields(cells[2])
			if len(metrics) < 2 {
				continue
			}
			pending = &common.GPU{
				Type:             "ascend",
				Index:            uint(deviceID),
				ProductName:      strings.Join(fields[1:], " "),
				PersistenceMode:  "N/A",
				DisplayActive:    "N/A",
				ECC:              "N/A",
				FanSpeed:         "N/A",
				Temperature:      ascendValueWithUnit(metrics[1], "C"),
				PerformanceState: strings.TrimSpace(cells[1]),
				PowerDraw:        ascendValueWithUnit(metrics[0], "W"),
				MaxPowerLimit:    "N/A",
				ComputeMode:      "N/A",
				MigMode:          "N/A",
			}
			continue
		}

		metrics := strings.Fields(cells[2])
		if len(metrics) == 0 {
			pending = nil
			continue
		}
		pending.BusID = strings.TrimSpace(cells[1])
		pending.GPUUtil = ascendValueWithUnit(metrics[0], "%")
		pending.MemUsed, pending.MemTotal = ascendMemoryUsage(cells[2])
		deviceIndexes[pending.Index] = len(info.GPUs)
		info.GPUs = append(info.GPUs, *pending)
		pending = nil
	}

	return info
}

func ascendTableCells(line string) []string {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "|") || !strings.HasSuffix(line, "|") {
		return nil
	}
	parts := strings.Split(line, "|")
	if len(parts) < 3 {
		return nil
	}
	cells := make([]string, 0, len(parts)-2)
	for _, part := range parts[1 : len(parts)-1] {
		cells = append(cells, strings.TrimSpace(part))
	}
	return cells
}

func ascendMemoryUsage(value string) (string, string) {
	matches := ascendMemoryPattern.FindAllStringSubmatch(value, -1)
	if len(matches) == 0 {
		return "N/A", "N/A"
	}
	// 910B exposes both generic memory and HBM. Prefer the last memory pool
	// with a non-zero capacity, which selects HBM while retaining compatibility
	// with devices that only expose Memory-Usage.
	selected := matches[len(matches)-1]
	for i := len(matches) - 1; i >= 0; i-- {
		if total, err := strconv.ParseFloat(matches[i][2], 64); err == nil && total > 0 {
			selected = matches[i]
			break
		}
	}
	return selected[1] + " MB", selected[2] + " MB"
}

func ascendValueWithUnit(value, unit string) string {
	value = strings.TrimSpace(value)
	if value == "" || strings.EqualFold(value, "N/A") || strings.EqualFold(value, "NA") {
		return "N/A"
	}
	return value + " " + unit
}
