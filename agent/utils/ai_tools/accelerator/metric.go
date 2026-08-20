package accelerator

import (
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/utils/re"
)

func metric(display, unit string) Metric {
	result := Metric{Display: display, Unit: unit}
	value := strings.TrimSpace(display)
	if value == "" || strings.EqualFold(value, "N/A") || strings.EqualFold(value, "NA") {
		return result
	}
	matched := re.GetRegex(re.AcceleratorMetricValuePattern).FindStringSubmatch(value)
	if len(matched) != 3 {
		return result
	}
	parsed, err := strconv.ParseFloat(matched[1], 64)
	if err != nil {
		return result
	}
	normalized, ok := convertMetricUnit(parsed, matched[2], unit)
	if !ok {
		return result
	}
	result.Value = &normalized
	return result
}

func memoryMetric(display string) Metric {
	return metric(display, "MiB")
}

func convertMetricUnit(value float64, sourceUnit, targetUnit string) (float64, bool) {
	source := normalizeUnit(sourceUnit)
	target := normalizeUnit(targetUnit)
	if source == "" {
		source = target
	}

	switch target {
	case "%":
		return value, source == "%" || source == "percent" || source == "pct"
	case "w":
		switch source {
		case "w", "watt", "watts":
			return value, true
		case "mw":
			return value / 1000, true
		case "kw":
			return value * 1000, true
		}
	case "°c":
		switch source {
		case "c", "°c", "℃":
			return value, true
		case "f", "°f", "℉":
			return (value - 32) * 5 / 9, true
		case "k":
			return value - 273.15, true
		}
	case "mhz":
		switch source {
		case "hz":
			return value / 1_000_000, true
		case "khz":
			return value / 1000, true
		case "mhz":
			return value, true
		case "ghz":
			return value * 1000, true
		}
	case "mib":
		switch source {
		case "b", "byte", "bytes":
			return value / (1024 * 1024), true
		case "kb":
			return value * 1000 / (1024 * 1024), true
		case "kib":
			return value / 1024, true
		case "mb":
			return value * 1_000_000 / (1024 * 1024), true
		case "mib":
			return value, true
		case "gb":
			return value * 1_000_000_000 / (1024 * 1024), true
		case "gib":
			return value * 1024, true
		case "tb":
			return value * 1_000_000_000_000 / (1024 * 1024), true
		case "tib":
			return value * 1024 * 1024, true
		}
	default:
		return value, source == target
	}
	return 0, false
}

func normalizeUnit(unit string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(unit)), ""))
}

func normalizedMemoryDisplay(display string) string {
	value := memoryMetric(display)
	if !value.Available() {
		return display
	}
	return strconv.FormatFloat(value.ValueOrZero(), 'f', -1, 64) + " MiB"
}
