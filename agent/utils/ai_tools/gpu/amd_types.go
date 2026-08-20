package gpu

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// amdSMIRow is one GPU entry from amd-smi JSON output. amd-smi has changed
// field names and nesting between releases, so the raw vendor response stays
// flexible and is normalized into Device in amd.go.
type amdSMIRow map[string]any

type amdSMIRows []amdSMIRow

// amdSMIMetricValue represents both the newer {"value": ..., "unit": ...}
// form and the scalar values returned by older amd-smi releases.
type amdSMIMetricValue struct {
	Value any
	Unit  string
}

func decodeAMDRows(data string) (amdSMIRows, error) {
	root, err := decodeAMDJSON(data)
	if err != nil {
		return nil, err
	}
	return amdRows(root), nil
}

func decodeAMDJSON(data string) (any, error) {
	for index, char := range data {
		if char != '{' && char != '[' {
			continue
		}
		decoder := json.NewDecoder(strings.NewReader(data[index:]))
		decoder.UseNumber()
		var value any
		if err := decoder.Decode(&value); err == nil {
			return value, nil
		}
	}
	return nil, fmt.Errorf("invalid JSON output")
}

func amdRows(value any) amdSMIRows {
	switch typed := value.(type) {
	case []any:
		rows := make(amdSMIRows, 0, len(typed))
		for _, item := range typed {
			if row, ok := item.(map[string]any); ok {
				rows = append(rows, amdSMIRow(row))
			}
		}
		return rows
	case map[string]any:
		row := amdSMIRow(typed)
		if _, ok := amdValueAt(row, "gpu"); ok {
			return amdSMIRows{row}
		}
		for _, key := range []string{"gpu_data", "data"} {
			if nested, ok := amdMapValue(typed, key); ok {
				if rows := amdRows(nested); len(rows) > 0 {
					return rows
				}
			}
		}
	}
	return nil
}

func amdObjectList(value any) amdSMIRows {
	switch typed := value.(type) {
	case []any:
		items := make(amdSMIRows, 0, len(typed))
		for _, item := range typed {
			if object, ok := item.(map[string]any); ok {
				items = append(items, amdSMIRow(object))
			}
		}
		return items
	case map[string]any:
		return amdSMIRows{amdSMIRow(typed)}
	default:
		return nil
	}
}

func amdGPUIndex(row amdSMIRow) (uint, bool) {
	value, ok := amdValueAt(row, "gpu")
	if !ok {
		return 0, false
	}
	number, err := strconv.ParseFloat(amdScalar(value), 64)
	if err != nil || number < 0 {
		return 0, false
	}
	return uint(number), true
}

func amdStringAt(row amdSMIRow, paths ...string) string {
	for _, path := range paths {
		if value, ok := amdValueAt(row, path); ok {
			if result := amdScalar(value); result != "" {
				return result
			}
		}
	}
	return ""
}

func amdMetricAt(row amdSMIRow, defaultUnit string, paths ...string) string {
	for _, path := range paths {
		value, ok := amdValueAt(row, path)
		if !ok {
			continue
		}
		if result := formatAMDMetric(value, defaultUnit); result != "" {
			return result
		}
	}
	return ""
}

func formatAMDMetric(value any, defaultUnit string) string {
	metric, ok := newAMDMetricValue(value)
	if !ok {
		return ""
	}
	formatted := amdScalar(metric.Value)
	if formatted == "" || strings.EqualFold(formatted, "N/A") {
		return formatted
	}
	if _, err := strconv.ParseFloat(formatted, 64); err != nil {
		return formatted
	}
	if metric.Unit == "" {
		metric.Unit = defaultUnit
	}
	if metric.Unit == "" {
		return formatted
	}
	return formatted + " " + metric.Unit
}

func newAMDMetricValue(value any) (amdSMIMetricValue, bool) {
	metric := amdSMIMetricValue{Value: value}
	object, ok := value.(map[string]any)
	if !ok {
		return metric, true
	}
	metric.Value, ok = amdMapValue(object, "value")
	if !ok {
		return amdSMIMetricValue{}, false
	}
	if unit, exists := amdMapValue(object, "unit"); exists {
		metric.Unit = amdScalar(unit)
	}
	return metric, true
}

func amdValueAt(row amdSMIRow, path string) (any, bool) {
	var current any = map[string]any(row)
	for _, key := range strings.Split(path, ".") {
		object, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = amdMapValue(object, key)
		if !ok {
			return nil, false
		}
	}
	return current, true
}

func amdMapValue(object map[string]any, key string) (any, bool) {
	wanted := normalizeAMDKey(key)
	for currentKey, value := range object {
		if normalizeAMDKey(currentKey) == wanted {
			return value, true
		}
	}
	return nil, false
}

func normalizeAMDKey(value string) string {
	return strings.Map(func(char rune) rune {
		if unicode.IsLetter(char) || unicode.IsDigit(char) {
			return unicode.ToLower(char)
		}
		return -1
	}, value)
}

func amdScalar(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case json.Number:
		return typed.String()
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(typed), 'f', -1, 32)
	case int:
		return strconv.Itoa(typed)
	case int64:
		return strconv.FormatInt(typed, 10)
	case uint:
		return strconv.FormatUint(uint64(typed), 10)
	case uint64:
		return strconv.FormatUint(typed, 10)
	default:
		return ""
	}
}
