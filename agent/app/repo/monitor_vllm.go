package repo

import (
	"context"
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
	Bucket            int64
	CompletedRequests *float64
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

func (r *VLLMMonitorRepo) Count(ctx context.Context, id uint, start, end time.Time) (int64, error) {
	var count int64
	db := global.VLLMMonitorDB.WithContext(ctx).Model(&model.MonitorVLLM{}).Where("app_install_id = ? AND created_at >= ? AND created_at <= ?", id, start, end)
	err := db.Count(&count).Error
	return count, err
}

func (r *VLLMMonitorRepo) History(ctx context.Context, id uint, start, end time.Time, seconds int64, aggregation, group string) ([]VLLMHistoryPoint, error) {
	db := global.VLLMMonitorDB.WithContext(ctx).Model(&model.MonitorVLLM{}).Where("app_install_id = ? AND created_at >= ? AND created_at <= ?", id, start, end)
	var metrics []string
	if group == "" || group == "overview" {
		metrics = append(metrics, "running", "waiting", "cache_usage", "prompt_throughput", "generation_throughput", "request_throughput")
		if seconds <= 0 || aggregation == "max" {
			metrics = append(metrics, "prefix_cache_hit_rate")
		}
	}
	if seconds <= 0 || aggregation == "max" {
		if group == "" || group == "latency" {
			metrics = append(metrics, "queue_time", "time_to_first_token", "time_per_output_token", "request_latency", "prefill_time", "decode_time", "time_to_first_token_p50", "time_to_first_token_p90", "time_to_first_token_p95", "time_to_first_token_p99", "time_per_output_token_p50", "time_per_output_token_p90", "time_per_output_token_p95", "time_per_output_token_p99", "request_latency_p50", "request_latency_p90", "request_latency_p95", "request_latency_p99")
		}
		if group == "" || group == "requests" {
			metrics = append(metrics, "average_generation_tokens")
		}
	}
	var columns []string
	if seconds > 0 {
		operation := "AVG"
		if aggregation == "max" {
			operation = "MAX"
		}
		columns = []string{fmt.Sprintf("(CAST(strftime('%%s', created_at) AS INTEGER) - %d) / %d AS bucket", start.Unix(), seconds)}
		for _, column := range metrics {
			columns = append(columns, operation+"("+column+") AS "+column)
		}
		db = db.Group("bucket").Order("bucket ASC")
	} else {
		columns = append([]string{"id", "created_at", "app_install_id", "status", "sample_interval_seconds"}, metrics...)
		if group == "" || group == "requests" {
			columns = append(columns, "histogram_deltas", "request_deltas")
		}
		db = db.Order("created_at ASC, id ASC")
	}
	if group == "" || group == "overview" {
		count := "CASE WHEN json_valid(request_deltas) THEN CASE WHEN json_extract(request_deltas, '$.seconds') > 0 THEN ROUND(request_throughput * json_extract(request_deltas, '$.seconds')) END END"
		if seconds > 0 {
			count = "SUM(" + count + ")"
		}
		columns = append(columns, count+" AS completed_requests")
	}
	var points []VLLMHistoryPoint
	err := db.Select(strings.Join(columns, ", ")).Scan(&points).Error
	return points, err
}

func (r *VLLMMonitorRepo) WalkDeltas(ctx context.Context, id uint, start, end time.Time, seconds int64, aggregation, group string, visit func(int64, string, string) error) error {
	histograms, requests := "''", "''"
	if group == "" || group == "requests" || (group == "latency" && aggregation != "max") {
		histograms = "COALESCE(histogram_deltas, '')"
	}
	if group == "" || group == "requests" || (group == "overview" && aggregation != "max") {
		requests = "COALESCE(request_deltas, '')"
	}
	if histograms == "''" && requests == "''" {
		return ctx.Err()
	}
	bucket := fmt.Sprintf("(CAST(strftime('%%s', created_at) AS INTEGER) - %d) / %d", start.Unix(), seconds)
	rows, err := global.VLLMMonitorDB.WithContext(ctx).Model(&model.MonitorVLLM{}).
		Where("app_install_id = ? AND created_at >= ? AND created_at <= ?", id, start, end).
		Select(bucket + ", " + histograms + ", " + requests).Rows()
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		if err := ctx.Err(); err != nil {
			return err
		}
		var index int64
		var histograms, requests string
		if err := rows.Scan(&index, &histograms, &requests); err != nil {
			return err
		}
		if err := visit(index, histograms, requests); err != nil {
			return err
		}
	}
	return rows.Err()
}
