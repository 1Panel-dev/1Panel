package repo

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
	"gorm.io/gorm"
)

type VLLMMonitorRepo struct{}

type VLLMHistoryPoint struct {
	model.MonitorVLLM
	Bucket           int64
	HistogramSamples string
}

func (r *VLLMMonitorRepo) Create(point *model.MonitorVLLM) error {
	return global.VLLMMonitorDB.Create(point).Error
}

func (r *VLLMMonitorRepo) Latest(id uint) (model.MonitorVLLM, error) {
	var point model.MonitorVLLM
	db := global.VLLMMonitorDB.Where("app_install_id = ?", id)
	err := db.Order("created_at DESC, id DESC").First(&point).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return point, nil
	}
	return point, err
}

func (r *VLLMMonitorRepo) CleanTarget(id uint) error {
	return global.VLLMMonitorDB.Where("app_install_id = ?", id).Delete(&model.MonitorVLLM{}).Error
}

func (r *VLLMMonitorRepo) DeleteBefore(before time.Time) error {
	return global.VLLMMonitorDB.Where("created_at < ?", before).Delete(&model.MonitorVLLM{}).Error
}

func (r *VLLMMonitorRepo) Count(id uint, start, end time.Time) (int64, error) {
	var count int64
	db := global.VLLMMonitorDB.Model(&model.MonitorVLLM{}).Where("app_install_id = ? AND created_at >= ? AND created_at <= ?", id, start, end)
	err := db.Count(&count).Error
	return count, err
}

func (r *VLLMMonitorRepo) History(id uint, start, end time.Time, seconds int64, aggregation string) ([]VLLMHistoryPoint, error) {
	db := global.VLLMMonitorDB.Model(&model.MonitorVLLM{}).Where("app_install_id = ? AND created_at >= ? AND created_at <= ?", id, start, end)
	metrics := []string{"running", "waiting", "cache_usage", "prompt_throughput", "generation_throughput", "request_throughput", "time_to_first_token", "time_per_output_token", "request_latency", "prefill_time", "decode_time", "time_to_first_token_p50", "time_to_first_token_p90", "time_to_first_token_p95", "time_to_first_token_p99", "time_per_output_token_p50", "time_per_output_token_p90", "time_per_output_token_p95", "time_per_output_token_p99", "request_latency_p50", "request_latency_p90", "request_latency_p95", "request_latency_p99"}
	var columns []string
	if seconds > 0 {
		operation := "AVG"
		if aggregation == "max" {
			operation = "MAX"
		}
		columns = []string{fmt.Sprintf("(CAST(strftime('%%s', created_at) AS INTEGER) - %d) / %d AS bucket", start.Unix(), seconds)}
		for _, column := range metrics {
			if aggregation != "max" && (strings.HasPrefix(column, "time_to_first_token_p") || strings.HasPrefix(column, "time_per_output_token_p") || strings.HasPrefix(column, "request_latency_p")) {
				continue
			}
			columns = append(columns, operation+"("+column+") AS "+column)
		}
		if aggregation != "max" {
			columns = append(columns, "json_group_array(json(NULLIF(histogram_deltas, ''))) AS histogram_samples")
		}
		db = db.Group("bucket").Order("bucket ASC")
	} else {
		columns = append([]string{"id", "created_at", "app_install_id", "status"}, metrics...)
		db = db.Order("created_at ASC, id ASC")
	}
	var points []VLLMHistoryPoint
	err := db.Select(strings.Join(columns, ", ")).Scan(&points).Error
	return points, err
}
