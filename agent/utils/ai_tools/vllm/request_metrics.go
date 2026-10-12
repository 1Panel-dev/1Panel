package vllm

import (
	"encoding/json"
	"strconv"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/utils/re"
)

type requestMetrics struct {
	Seconds       float64            `json:"seconds"`
	PrefixHits    *float64           `json:"prefixHits,omitempty"`
	PrefixQueries *float64           `json:"prefixQueries,omitempty"`
	FinishReasons map[string]float64 `json:"finishReasons,omitempty"`
}

func calculateRequestMetrics(current, previous Metrics, elapsed float64, result *model.MonitorVLLM) error {
	sample := requestMetrics{Seconds: elapsed, FinishReasons: map[string]float64{}}
	hits := counterDelta(current["prefix_cache_hits_total"], previous["prefix_cache_hits_total"])
	queries := counterDelta(current["prefix_cache_queries_total"], previous["prefix_cache_queries_total"])
	if hits != nil && queries != nil && *hits <= *queries {
		sample.PrefixHits, sample.PrefixQueries = hits, queries
		if *queries > 0 {
			rate := *hits / *queries * 100
			result.PrefixCacheHitRate = &rate
		}
	}
	before := groupFinishReasons(previous["request_success_total"])
	for reason, values := range groupFinishReasons(current["request_success_total"]) {
		if delta := counterDelta(values, before[reason]); delta != nil {
			sample.FinishReasons[reason] = *delta
			if result.FinishReasons == nil {
				result.FinishReasons = map[string]float64{}
			}
			result.FinishReasons[reason] = *delta / elapsed
		}
	}
	encoded, err := json.Marshal(sample)
	if err != nil {
		return err
	}
	result.RequestDeltas = string(encoded)
	return nil
}

func groupFinishReasons(samples map[string]float64) map[string]map[string]float64 {
	groups := map[string]map[string]float64{}
	for labels, value := range samples {
		for _, label := range re.GetRegex(re.VLLMMetricLabelPattern).FindAllStringSubmatch(labels, -1) {
			if label[1] != "finished_reason" {
				continue
			}
			reason, err := strconv.Unquote(label[2])
			if err != nil || reason == "" {
				continue
			}
			if groups[reason] == nil {
				groups[reason] = map[string]float64{}
			}
			groups[reason][labels] = value
		}
	}
	return groups
}
