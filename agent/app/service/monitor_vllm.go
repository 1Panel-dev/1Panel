package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/ai_tools/vllm"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/robfig/cron/v3"
	"golang.org/x/sync/errgroup"
)

var vllmMonitorMutex sync.Mutex
var vllmMetricsClient = &http.Client{
	Timeout:       5 * time.Second,
	Transport:     &http.Transport{DialContext: (&net.Dialer{Timeout: 3 * time.Second}).DialContext, IdleConnTimeout: 30 * time.Second},
	CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
}

func (m *MonitorService) LoadVLLMMonitorData(req dto.MonitorVLLMSearch) (dto.MonitorVLLMData, error) {
	data := dto.MonitorVLLMData{Points: []model.MonitorVLLM{}}
	if !req.EndTime.After(req.StartTime) {
		return data, fmt.Errorf("invalid vLLM history request")
	}
	install, err := appInstallRepo.GetFirst(repo.WithByID(req.AppInstallID))
	if err != nil {
		return data, err
	}
	if install.App.Key != "vllm" {
		return data, fmt.Errorf("not a vLLM installation")
	}

	loc, err := time.LoadLocation(common.LoadTimeZoneByCmd())
	if err != nil {
		return data, err
	}
	req.StartTime, req.EndTime = req.StartTime.In(loc), req.EndTime.In(loc)
	data.SampleCount, err = vllmMonitorRepo.Count(req.AppInstallID, req.StartTime, req.EndTime)
	if err != nil || data.SampleCount == 0 {
		return data, err
	}
	if data.SampleCount > 1200 {
		data.BucketSeconds = (req.EndTime.Unix() - req.StartTime.Unix() + 600) / 600
	}
	points, err := vllmMonitorRepo.History(req.AppInstallID, req.StartTime, req.EndTime, data.BucketSeconds, req.Aggregation)
	if err != nil {
		return data, err
	}
	if data.BucketSeconds > 0 {
		if req.Aggregation != "max" {
			for i := range points {
				if err := vllm.AggregateHistograms(points[i].HistogramSamples, &points[i].MonitorVLLM); err != nil {
					return data, err
				}
			}
		}
		next := 0
		for bucket := int64(0); bucket <= (req.EndTime.Unix()-req.StartTime.Unix())/data.BucketSeconds; bucket++ {
			var point model.MonitorVLLM
			if next < len(points) && points[next].Bucket == bucket {
				point = points[next].MonitorVLLM
				next++
			}
			point.AppInstallID = req.AppInstallID
			point.CreatedAt = time.Unix(req.StartTime.Unix()+bucket*data.BucketSeconds, 0).In(loc)
			if bucket == 0 {
				point.CreatedAt = req.StartTime
			}
			data.Points = append(data.Points, point)
		}
	} else {
		setting, err := settingRepo.GetValueByKey("VLLMMonitorInterval")
		if err != nil {
			return data, err
		}
		intervalSeconds, err := strconv.Atoi(setting)
		if err != nil || intervalSeconds <= 0 {
			return data, fmt.Errorf("invalid vLLM monitoring interval: %s", setting)
		}
		interval := time.Duration(intervalSeconds) * time.Second
		for i, point := range points {
			if i > 0 && point.CreatedAt.Sub(points[i-1].CreatedAt) > 2*interval {
				data.Points = append(data.Points, model.MonitorVLLM{CreatedAt: points[i-1].CreatedAt.Add(interval), AppInstallID: req.AppInstallID})
			}
			data.Points = append(data.Points, point.MonitorVLLM)
		}
	}
	return data, nil
}

func (m *MonitorService) LoadVLLMCurrent(ctx context.Context, req dto.MonitorVLLMCurrent) (model.MonitorVLLM, error) {
	install, err := appInstallRepo.GetFirst(repo.WithByID(req.AppInstallID))
	if err != nil {
		return model.MonitorVLLM{}, err
	}
	if install.App.Key != "vllm" {
		return model.MonitorVLLM{}, fmt.Errorf("not a vLLM installation")
	}
	previous, err := collectVLLMMetrics(ctx, install, model.MonitorVLLM{})
	if err != nil {
		global.LOG.Debugf("Collect vLLM metrics on port %d failed: %v", install.HttpPort, err)
		return previous, nil
	}
	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return model.MonitorVLLM{}, ctx.Err()
	case <-timer.C:
	}
	point, err := collectVLLMMetrics(ctx, install, previous)
	if err != nil {
		global.LOG.Debugf("Collect vLLM metrics on port %d failed: %v", install.HttpPort, err)
	}
	return point, nil
}

func (m *MonitorService) CleanVLLMMonitor(req dto.MonitorVLLMClean) error {
	install, err := appInstallRepo.GetFirst(repo.WithByID(req.AppInstallID))
	if err != nil {
		return err
	}
	if install.App.Key != "vllm" {
		return fmt.Errorf("not a vLLM installation")
	}
	vllmMonitorMutex.Lock()
	defer vllmMonitorMutex.Unlock()
	return vllmMonitorRepo.CleanTarget(req.AppInstallID)
}

func StartVLLMMonitor(interval string) error {
	seconds, err := strconv.Atoi(interval)
	if err != nil || seconds < 10 || seconds > 43200 {
		return fmt.Errorf("invalid vLLM monitoring interval: %s", interval)
	}
	service := &MonitorService{}
	job := cron.NewChain(cron.Recover(cron.DefaultLogger)).Then(cron.FuncJob(service.saveVLLMData))
	id, err := global.Cron.AddFunc(fmt.Sprintf("@every %ds", seconds), func() { go job.Run() })
	if err != nil {
		return err
	}
	if global.VLLMMonitorCronID != 0 {
		global.Cron.Remove(global.VLLMMonitorCronID)
	}
	global.VLLMMonitorCronID = id
	return nil
}

func (m *MonitorService) saveVLLMData() {
	if !vllmMonitorMutex.TryLock() {
		return
	}
	defer vllmMonitorMutex.Unlock()
	status, err := settingRepo.GetValueByKey("VLLMMonitorStatus")
	if err != nil {
		global.LOG.Errorf("Load vLLM monitoring status failed: %v", err)
		return
	}
	if status != constant.StatusEnable {
		return
	}
	retention, err := settingRepo.GetValueByKey("VLLMMonitorStoreDays")
	if err != nil {
		global.LOG.Errorf("Load vLLM monitoring retention failed: %v", err)
		return
	}
	days, err := strconv.Atoi(retention)
	if err != nil || days < 1 {
		global.LOG.Errorf("Invalid vLLM monitoring retention: %s", retention)
		return
	}
	if err := vllmMonitorRepo.DeleteBefore(time.Now().AddDate(0, 0, -days)); err != nil {
		global.LOG.Errorf("Clean vLLM monitoring data failed: %v", err)
	}
	apps, err := appRepo.GetBy(appRepo.WithKey("vllm"))
	if err != nil {
		global.LOG.Errorf("Load vLLM apps for monitoring failed: %v", err)
		return
	}
	if len(apps) == 0 {
		return
	}
	ids := make([]uint, 0, len(apps))
	for _, app := range apps {
		ids = append(ids, app.ID)
	}
	installs, err := appInstallRepo.ListBy(context.Background(), appInstallRepo.WithAppIdsIn(ids))
	if err != nil {
		global.LOG.Errorf("Load vLLM instances for monitoring failed: %v", err)
		return
	}
	var group errgroup.Group
	group.SetLimit(4)
	for _, install := range installs {
		group.Go(func() error {
			previous, err := vllmMonitorRepo.Latest(install.ID)
			if err != nil {
				return err
			}
			point, err := collectVLLMMetrics(context.Background(), install, previous)
			if err != nil {
				global.LOG.Debugf("Collect vLLM metrics for %s failed: %v", install.Name, err)
			}
			return vllmMonitorRepo.Create(&point)
		})
	}
	if err := group.Wait(); err != nil {
		global.LOG.Errorf("Save vLLM monitoring data failed: %v", err)
	}
}

func collectVLLMMetrics(ctx context.Context, install model.AppInstall, previous model.MonitorVLLM) (model.MonitorVLLM, error) {
	point := model.MonitorVLLM{AppInstallID: install.ID, CreatedAt: time.Now(), Status: "unavailable"}
	if install.HttpPort <= 0 || install.HttpPort > 65535 {
		return point, fmt.Errorf("invalid vLLM port")
	}
	var env map[string]interface{}
	if err := json.Unmarshal([]byte(install.Env), &env); err != nil {
		return point, err
	}
	host, _ := env[constant.HostIP].(string)
	host = strings.Trim(host, "[]")
	switch host {
	case "", "0.0.0.0":
		host = "127.0.0.1"
	case "::":
		host = "::1"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+net.JoinHostPort(host, strconv.Itoa(install.HttpPort))+"/metrics", nil)
	if err != nil {
		return point, err
	}
	response, err := vllmMetricsClient.Do(req)
	if err != nil {
		return point, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return point, fmt.Errorf("vLLM metrics returned HTTP %d", response.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, 8*1024*1024+1))
	if err != nil {
		return point, err
	}
	if len(body) > 8*1024*1024 {
		return point, fmt.Errorf("vLLM metrics response too large")
	}
	metrics, err := vllm.Parse(bytes.NewReader(body))
	if err != nil {
		return point, err
	}
	point.CreatedAt = time.Now()
	raw, err := json.Marshal(metrics)
	if err != nil {
		return point, err
	}
	var before vllm.Metrics
	elapsed := point.CreatedAt.Sub(previous.CreatedAt).Seconds()
	if previous.Status == "ok" && elapsed > 0 {
		if err := json.Unmarshal([]byte(previous.RawMetrics), &before); err != nil {
			return point, err
		}
	}
	calculated, err := vllm.Calculate(metrics, before, elapsed)
	if err != nil {
		return point, err
	}
	calculated.AppInstallID = install.ID
	calculated.CreatedAt = point.CreatedAt
	calculated.RawMetrics = string(raw)
	calculated.Status = "ok"
	return calculated, nil
}
