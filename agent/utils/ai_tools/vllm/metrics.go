package vllm

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/utils/re"
)

type Metrics map[string]map[string]float64

type histogram struct {
	Sum     float64            `json:"sum"`
	Count   float64            `json:"count"`
	Buckets map[string]float64 `json:"buckets,omitempty"`
}

func Parse(reader io.Reader) (Metrics, error) {
	metrics := Metrics{}
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "vllm:") {
			continue
		}
		match := re.GetRegex(re.VLLMMetricSamplePattern).FindStringSubmatch(line)
		if match == nil {
			return nil, fmt.Errorf("invalid vLLM metric sample")
		}
		name := strings.TrimPrefix(match[1], "vllm:")
		switch name {
		case "num_requests_running", "num_requests_waiting", "kv_cache_usage_perc", "gpu_cache_usage_perc", "prompt_tokens_total", "generation_tokens_total", "request_success_total":
		default:
			base := strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(name, "_sum"), "_count"), "_bucket")
			if base == name {
				continue
			}
			switch base {
			case "time_to_first_token_seconds", "time_per_output_token_seconds", "inter_token_latency_seconds", "e2e_request_latency_seconds", "request_prefill_time_seconds", "request_decode_time_seconds":
			default:
				continue
			}
		}
		value, err := strconv.ParseFloat(match[3], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid vLLM metric value: %w", err)
		}
		if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
			continue
		}
		if metrics[name] == nil {
			metrics[name] = map[string]float64{}
		}
		labels := re.GetRegex(re.VLLMMetricLabelPattern).FindAllStringSubmatch(match[2], -1)
		parts := make([]string, 0, len(labels))
		for _, label := range labels {
			parts = append(parts, label[1]+"="+label[2])
		}
		sort.Strings(parts)
		metrics[name]["{"+strings.Join(parts, ",")+"}"] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(metrics) == 0 {
		return nil, fmt.Errorf("no supported vLLM metrics")
	}
	return metrics, nil
}

func Calculate(current, previous Metrics, elapsed float64) (model.MonitorVLLM, error) {
	result := model.MonitorVLLM{}
	for name, target := range map[string]**float64{
		"num_requests_running": &result.Running,
		"num_requests_waiting": &result.Waiting,
	} {
		if len(current[name]) == 0 {
			continue
		}
		value := 0.0
		for _, sample := range current[name] {
			value += sample
		}
		*target = &value
	}
	cache := current["kv_cache_usage_perc"]
	if len(cache) == 0 {
		cache = current["gpu_cache_usage_perc"]
	}
	if len(cache) > 0 {
		value := 0.0
		for _, sample := range cache {
			value += sample
		}
		value = value / float64(len(cache)) * 100
		result.CacheUsage = &value
	}
	if elapsed <= 0 {
		return result, nil
	}
	for name, target := range map[string]**float64{
		"prompt_tokens_total":     &result.PromptThroughput,
		"generation_tokens_total": &result.GenerationThroughput,
		"request_success_total":   &result.RequestThroughput,
	} {
		if value := counterDelta(current[name], previous[name]); value != nil {
			rate := *value / elapsed
			*target = &rate
		}
	}
	histograms := map[string]histogram{}
	for _, name := range []string{"time_to_first_token_seconds", "inter_token_latency_seconds", "e2e_request_latency_seconds", "request_prefill_time_seconds", "request_decode_time_seconds"} {
		source := name
		if name == "inter_token_latency_seconds" && len(current[name+"_count"]) == 0 {
			source = "time_per_output_token_seconds"
		}
		sum := counterDelta(current[source+"_sum"], previous[source+"_sum"])
		count := counterDelta(current[source+"_count"], previous[source+"_count"])
		if sum == nil || count == nil {
			continue
		}
		histograms[name] = histogram{Sum: *sum, Count: *count, Buckets: bucketDeltas(current[source+"_bucket"], previous[source+"_bucket"], *count)}
	}
	applyHistograms(&result, histograms)
	encoded, err := json.Marshal(histograms)
	if err != nil {
		return model.MonitorVLLM{}, err
	}
	result.HistogramDeltas = string(encoded)
	return result, nil
}

func AggregateHistograms(encoded string, result *model.MonitorVLLM) error {
	var samples []map[string]histogram
	if err := json.Unmarshal([]byte(encoded), &samples); err != nil {
		return err
	}
	merged := map[string]histogram{}
	for _, sample := range samples {
		for name, item := range sample {
			if item.Count == 0 {
				continue
			}
			total, exists := merged[name]
			if !exists {
				merged[name] = item
				continue
			}
			total.Sum += item.Sum
			total.Count += item.Count
			if len(total.Buckets) != len(item.Buckets) {
				total.Buckets = nil
			}
			for bound, count := range total.Buckets {
				addition, ok := item.Buckets[bound]
				if !ok {
					total.Buckets = nil
					break
				}
				total.Buckets[bound] = count + addition
			}
			merged[name] = total
		}
	}
	applyHistograms(result, merged)
	return nil
}

func applyHistograms(result *model.MonitorVLLM, histograms map[string]histogram) {
	for _, item := range []struct {
		name      string
		mean      **float64
		quantiles []**float64
	}{
		{"time_to_first_token_seconds", &result.TimeToFirstToken, []**float64{&result.TimeToFirstTokenP50, &result.TimeToFirstTokenP90, &result.TimeToFirstTokenP95, &result.TimeToFirstTokenP99}},
		{"inter_token_latency_seconds", &result.TimePerOutputToken, []**float64{&result.TimePerOutputTokenP50, &result.TimePerOutputTokenP90, &result.TimePerOutputTokenP95, &result.TimePerOutputTokenP99}},
		{"e2e_request_latency_seconds", &result.RequestLatency, []**float64{&result.RequestLatencyP50, &result.RequestLatencyP90, &result.RequestLatencyP95, &result.RequestLatencyP99}},
		{"request_prefill_time_seconds", &result.PrefillTime, nil},
		{"request_decode_time_seconds", &result.DecodeTime, nil},
	} {
		histogram, ok := histograms[item.name]
		if !ok || histogram.Count <= 0 {
			continue
		}
		mean := histogram.Sum / histogram.Count
		*item.mean = &mean
		if len(item.quantiles) == 0 {
			continue
		}
		quantiles := histogramQuantiles(histogram.Buckets)
		for i, target := range item.quantiles {
			*target = quantiles[i]
		}
	}
}

func counterDelta(current, previous map[string]float64) *float64 {
	if len(current) == 0 || len(current) != len(previous) {
		return nil
	}
	value := 0.0
	for label, sample := range current {
		before, ok := previous[label]
		if !ok || sample < before {
			return nil
		}
		value += sample - before
	}
	if math.IsInf(value, 0) || math.IsNaN(value) {
		return nil
	}
	return &value
}

func bucketDeltas(current, previous map[string]float64, count float64) map[string]float64 {
	if len(current) == 0 || len(current) != len(previous) {
		return nil
	}
	groups := map[string]map[string]float64{}
	for labels, sample := range current {
		before, ok := previous[labels]
		if !ok || sample < before {
			return nil
		}
		bound := ""
		var identity []string
		for _, label := range re.GetRegex(re.VLLMMetricLabelPattern).FindAllStringSubmatch(labels, -1) {
			if label[1] == "le" {
				raw, err := strconv.Unquote(label[2])
				if err != nil {
					return nil
				}
				upper, err := strconv.ParseFloat(raw, 64)
				if err != nil || math.IsNaN(upper) || upper < 0 {
					return nil
				}
				bound = strconv.FormatFloat(upper, 'g', -1, 64)
			} else {
				identity = append(identity, label[1]+"="+label[2])
			}
		}
		if bound == "" {
			return nil
		}
		group := strings.Join(identity, ",")
		if groups[group] == nil {
			groups[group] = map[string]float64{}
		}
		if _, exists := groups[group][bound]; exists {
			return nil
		}
		groups[group][bound] = sample - before
	}
	result := map[string]float64{}
	for _, buckets := range groups {
		if _, ok := buckets["+Inf"]; !ok || len(buckets) < 2 {
			return nil
		}
		if len(result) > 0 && len(result) != len(buckets) {
			return nil
		}
		for bound := range result {
			if _, ok := buckets[bound]; !ok {
				return nil
			}
		}
		bounds := make([]float64, 0, len(buckets))
		for bound := range buckets {
			value, _ := strconv.ParseFloat(bound, 64)
			bounds = append(bounds, value)
		}
		sort.Float64s(bounds)
		previousCount := 0.0
		for _, bound := range bounds {
			key := strconv.FormatFloat(bound, 'g', -1, 64)
			value := buckets[key]
			if value < previousCount {
				if previousCount-value > 1e-12*(previousCount+value) {
					return nil
				}
				value = previousCount
			}
			result[key] += value
			previousCount = value
		}
	}
	if math.Abs(result["+Inf"]-count) > 1e-12*(result["+Inf"]+count) {
		return nil
	}
	return result
}

func histogramQuantiles(buckets map[string]float64) [4]*float64 {
	var quantiles [4]*float64
	if len(buckets) < 2 || buckets["+Inf"] <= 0 {
		return quantiles
	}
	bounds := make([]float64, 0, len(buckets)-1)
	for bound := range buckets {
		value, err := strconv.ParseFloat(bound, 64)
		if err != nil || math.IsNaN(value) || value < 0 {
			return quantiles
		}
		if !math.IsInf(value, 1) {
			bounds = append(bounds, value)
		}
	}
	if len(bounds) == 0 {
		return quantiles
	}
	sort.Float64s(bounds)
	for i, q := range [...]float64{0.5, 0.9, 0.95, 0.99} {
		rank := q * buckets["+Inf"]
		lower, before := 0.0, 0.0
		value := bounds[len(bounds)-1]
		for _, upper := range bounds {
			count := buckets[strconv.FormatFloat(upper, 'g', -1, 64)]
			if count >= rank && count > before {
				value = lower + (upper-lower)*(rank-before)/(count-before)
				break
			}
			lower, before = upper, count
		}
		quantiles[i] = &value
	}
	return quantiles
}
