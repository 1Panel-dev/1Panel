package npu

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/re"
)

const ascendSMICommand = "npu-smi"

type Client struct{}

func New() (bool, Client) {
	return cmd.Which(ascendSMICommand), Client{}
}

type ascendDeviceKey struct {
	npuID  uint
	chipID uint
}

func (c Client) LoadInfo() (*Info, error) {
	return c.LoadInfoContext(context.Background())
}

func (c Client) LoadInfoContext(ctx context.Context) (*Info, error) {
	cmdMgr := cmd.NewCommandMgr(cmd.WithContext(ctx), cmd.WithTimeout(5*time.Second))
	itemData, err := cmdMgr.RunWithStdout(ascendSMICommand, "info")
	if err != nil {
		return nil, fmt.Errorf("calling %s failed: %w", ascendSMICommand, err)
	}
	return parseAscendSMI(itemData), nil
}

func parseAscendSMI(data string) *Info {
	info := &Info{Type: "ascend"}
	if match := re.GetRegex(re.AscendVersionPattern).FindStringSubmatch(data); len(match) == 2 {
		info.DriverVersion = match[1]
	}

	processSection := false
	deviceIndexes := make(map[ascendDeviceKey]int)
	chipMetricsHeader := ""
	var pending *Device

	for _, line := range strings.Split(data, "\n") {
		if strings.Contains(line, "Process id") && strings.Contains(line, "Process memory") {
			processSection = true
			pending = nil
			continue
		}

		cells := ascendTableCells(line)
		if len(cells) == 3 && strings.Contains(strings.ToUpper(cells[2]), "AICORE") {
			chipMetricsHeader = cells[2]
			continue
		}
		if processSection {
			if len(cells) != 4 && len(cells) != 5 {
				continue
			}
			deviceFields := strings.Fields(cells[0])
			if len(deviceFields) == 0 {
				continue
			}
			npuID, err := strconv.ParseUint(deviceFields[0], 10, 64)
			if err != nil {
				continue
			}
			var chipID uint64
			if len(deviceFields) > 1 {
				chipID, err = strconv.ParseUint(deviceFields[1], 10, 64)
				if err != nil {
					continue
				}
			}
			index, ok := deviceIndexes[ascendDeviceKey{npuID: uint(npuID), chipID: uint(chipID)}]
			if !ok {
				continue
			}
			processOffset := len(cells) - 3
			info.Devices[index].Processes = append(info.Devices[index].Processes, Process{
				PID:         strings.TrimSpace(cells[processOffset]),
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
			pending = &Device{
				Type:        "ascend",
				Index:       uint(deviceID),
				NPUIndex:    uint(deviceID),
				ProductName: strings.Join(fields[1:], " "),
				Temperature: ascendValueWithUnit(metrics[1], "C"),
				Health:      strings.TrimSpace(cells[1]),
				PowerDraw:   ascendValueWithUnit(metrics[0], "W"),
			}
			if usage := ascendUsageValues(cells[2]); len(usage) > 0 {
				pending.HugepagesUsed = usage[len(usage)-1][0]
				pending.HugepagesTotal = usage[len(usage)-1][1]
			}
			continue
		}

		deviceFields := strings.Fields(cells[0])
		if len(deviceFields) == 0 {
			pending = nil
			continue
		}
		chipID, err := strconv.ParseUint(deviceFields[0], 10, 64)
		if err != nil {
			pending = nil
			continue
		}
		pending.ChipIndex = uint(chipID)
		if len(deviceFields) > 1 {
			deviceID, err := strconv.ParseUint(deviceFields[1], 10, 64)
			if err != nil {
				pending = nil
				continue
			}
			pending.Index = uint(deviceID)
		}

		metrics := strings.Fields(cells[2])
		if len(metrics) == 0 {
			pending = nil
			continue
		}
		pending.BusID = strings.TrimSpace(cells[1])
		pending.AICore = ascendValueWithUnit(metrics[0], "%")
		pending.MemoryUsed, pending.MemoryTotal, pending.HBMUsed, pending.HBMTotal = ascendMemoryPools(cells[2], chipMetricsHeader)
		pending.MemUsed, pending.MemTotal = ascendMemoryUsage(cells[2])
		deviceIndexes[ascendDeviceKey{npuID: pending.NPUIndex, chipID: pending.ChipIndex}] = len(info.Devices)
		info.Devices = append(info.Devices, *pending)
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
	usage := ascendUsageValues(value)
	if len(usage) == 0 {
		return "N/A", "N/A"
	}
	// 910B exposes both generic memory and HBM. Prefer the last memory pool
	// with a non-zero capacity, which selects HBM while retaining compatibility
	// with devices that only expose Memory-Usage.
	selected := usage[len(usage)-1]
	for i := len(usage) - 1; i >= 0; i-- {
		if total, err := strconv.ParseFloat(usage[i][1], 64); err == nil && total > 0 {
			selected = usage[i]
			break
		}
	}
	return selected[0] + " MB", selected[1] + " MB"
}

func ascendMemoryPools(value, header string) (string, string, string, string) {
	usage := ascendUsageValues(value)
	if len(usage) == 0 {
		return "", "", "", ""
	}

	memoryUsed, memoryTotal := usage[0][0]+" MB", usage[0][1]+" MB"
	normalizedHeader := strings.NewReplacer("-", "", "_", "", " ", "").Replace(strings.ToUpper(header))
	if !strings.Contains(normalizedHeader, "HBMUSAGE") {
		return memoryUsed, memoryTotal, "", ""
	}

	hbm := usage[len(usage)-1]
	return memoryUsed, memoryTotal, hbm[0] + " MB", hbm[1] + " MB"
}

func ascendUsageValues(value string) [][2]string {
	matches := re.GetRegex(re.AscendMemoryPattern).FindAllStringSubmatch(value, -1)
	usage := make([][2]string, 0, len(matches))
	for _, match := range matches {
		usage = append(usage, [2]string{match[1], match[2]})
	}
	return usage
}

func ascendValueWithUnit(value, unit string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if strings.EqualFold(value, "N/A") || strings.EqualFold(value, "NA") {
		return "N/A"
	}
	if strings.HasSuffix(strings.ToUpper(value), strings.ToUpper(unit)) {
		return value
	}
	return value + " " + unit
}
