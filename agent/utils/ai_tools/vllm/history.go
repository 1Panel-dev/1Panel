package vllm

import (
	"encoding/json"
	"math"

	"github.com/1Panel-dev/1Panel/agent/app/model"
)

type HistoryAccumulator struct {
	histograms    map[string]histogram
	prefixHits    float64
	prefixQueries float64
	finishReasons map[string]float64
	finishSeconds map[string]float64
}

func (a *HistoryAccumulator) Add(histograms, requests, aggregation, group string) error {
	if histograms != "" {
		var sample map[string]histogram
		if aggregation == "max" || group == "requests" {
			var distributions struct {
				Prompt     *histogram `json:"request_prompt_tokens"`
				Generation *histogram `json:"request_generation_tokens"`
			}
			if err := json.Unmarshal([]byte(histograms), &distributions); err != nil {
				return err
			}
			sample = map[string]histogram{}
			if distributions.Prompt != nil {
				sample["request_prompt_tokens"] = *distributions.Prompt
			}
			if distributions.Generation != nil {
				sample["request_generation_tokens"] = *distributions.Generation
			}
		} else if group == "latency" {
			var latency struct {
				FirstToken *histogram `json:"time_to_first_token_seconds"`
				InterToken *histogram `json:"inter_token_latency_seconds"`
				Request    *histogram `json:"e2e_request_latency_seconds"`
				Prefill    *histogram `json:"request_prefill_time_seconds"`
				Decode     *histogram `json:"request_decode_time_seconds"`
				Queue      *histogram `json:"request_queue_time_seconds"`
			}
			if err := json.Unmarshal([]byte(histograms), &latency); err != nil {
				return err
			}
			sample = map[string]histogram{}
			for name, item := range map[string]*histogram{
				"time_to_first_token_seconds":  latency.FirstToken,
				"inter_token_latency_seconds":  latency.InterToken,
				"e2e_request_latency_seconds":  latency.Request,
				"request_prefill_time_seconds": latency.Prefill,
				"request_decode_time_seconds":  latency.Decode,
				"request_queue_time_seconds":   latency.Queue,
			} {
				if item != nil {
					sample[name] = *item
				}
			}
		} else if err := json.Unmarshal([]byte(histograms), &sample); err != nil {
			return err
		}
		if a.histograms == nil {
			a.histograms = map[string]histogram{}
		}
		for name, item := range sample {
			switch name {
			case "request_prefill_time_seconds", "request_decode_time_seconds", "request_queue_time_seconds":
				item.Buckets = nil
			}
			total, exists := a.histograms[name]
			if !exists {
				a.histograms[name] = item
				continue
			}
			total.Seconds += item.Seconds
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
			a.histograms[name] = total
		}
	}
	if requests == "" {
		return nil
	}
	if group == "overview" {
		var prefix struct {
			Seconds float64  `json:"seconds"`
			Hits    *float64 `json:"prefixHits"`
			Queries *float64 `json:"prefixQueries"`
		}
		if err := json.Unmarshal([]byte(requests), &prefix); err != nil {
			return err
		}
		if prefix.Seconds > 0 && prefix.Hits != nil && prefix.Queries != nil {
			a.prefixHits += *prefix.Hits
			a.prefixQueries += *prefix.Queries
		}
		return nil
	}
	var sample requestMetrics
	if err := json.Unmarshal([]byte(requests), &sample); err != nil {
		return err
	}
	if sample.Seconds <= 0 {
		return nil
	}
	if group == "" && aggregation != "max" && sample.PrefixHits != nil && sample.PrefixQueries != nil {
		a.prefixHits += *sample.PrefixHits
		a.prefixQueries += *sample.PrefixQueries
	}
	for reason, count := range sample.FinishReasons {
		if a.finishReasons == nil {
			a.finishReasons = map[string]float64{}
		}
		if aggregation == "max" {
			a.finishReasons[reason] = math.Max(a.finishReasons[reason], count/sample.Seconds)
		} else {
			if a.finishSeconds == nil {
				a.finishSeconds = map[string]float64{}
			}
			a.finishReasons[reason] += count
			a.finishSeconds[reason] += sample.Seconds
		}
	}
	return nil
}

func (a *HistoryAccumulator) Apply(result *model.MonitorVLLM, aggregation string) {
	applyHistograms(result, a.histograms, aggregation != "max")
	for reason, count := range a.finishReasons {
		if result.FinishReasons == nil {
			result.FinishReasons = map[string]float64{}
		}
		if aggregation == "max" {
			result.FinishReasons[reason] = count
		} else {
			result.FinishReasons[reason] = count / a.finishSeconds[reason]
		}
	}
	if aggregation != "max" && a.prefixQueries > 0 {
		rate := a.prefixHits / a.prefixQueries * 100
		result.PrefixCacheHitRate = &rate
	}
}
