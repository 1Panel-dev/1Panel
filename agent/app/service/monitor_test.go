package service

import (
	"io"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/re"
	"github.com/glebarez/sqlite"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func setupMonitorDBs(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	cfg := &gorm.Config{Logger: gormlogger.Default.LogMode(gormlogger.Silent)}

	monitorDB, err := gorm.Open(sqlite.Open(filepath.Join(dir, "monitor.db")), cfg)
	if err != nil {
		t.Fatalf("open monitor db: %v", err)
	}
	gpuDB, err := gorm.Open(sqlite.Open(filepath.Join(dir, "gpu_monitor.db")), cfg)
	if err != nil {
		t.Fatalf("open gpu monitor db: %v", err)
	}
	if err := monitorDB.AutoMigrate(&model.MonitorBase{}, &model.MonitorNetwork{}, &model.MonitorIO{}); err != nil {
		t.Fatalf("migrate monitor db: %v", err)
	}
	if err := gpuDB.AutoMigrate(&model.MonitorGPU{}); err != nil {
		t.Fatalf("migrate gpu monitor db: %v", err)
	}
	for _, stmt := range []string{
		"CREATE INDEX IF NOT EXISTS idx_monitor_ios_name_created ON monitor_ios(name, created_at)",
		"CREATE INDEX IF NOT EXISTS idx_monitor_networks_name_created ON monitor_networks(name, created_at)",
	} {
		if err := monitorDB.Exec(stmt).Error; err != nil {
			t.Fatalf("create index: %v", err)
		}
	}
	if err := gpuDB.Exec("CREATE INDEX IF NOT EXISTS idx_monitor_gpus_product_created ON monitor_gpus(product_name, created_at)").Error; err != nil {
		t.Fatalf("create gpu index: %v", err)
	}
	global.MonitorDB = monitorDB
	global.GPUMonitorDB = gpuDB

	// production binaries run these during server init; tests must do it explicitly
	re.Init()
	if global.LOG == nil {
		global.LOG = logrus.New()
		global.LOG.SetOutput(io.Discard)
	}
}

func seedMonitorRows(t *testing.T) time.Time {
	t.Helper()
	// sqlite stores timestamps as RFC3339 text and the service converts query bounds to the
	// local zone, so seeds must use local time for range comparisons to match
	base := time.Now().Add(-time.Minute).Truncate(time.Second)
	ioRows := []model.MonitorIO{
		{BaseModel: model.BaseModel{CreatedAt: base}, Name: "all", Read: 1, Write: 1},
		{BaseModel: model.BaseModel{CreatedAt: base}, Name: "sda", Read: 2, Write: 3},
		{BaseModel: model.BaseModel{CreatedAt: base}, Name: "veth-old", Read: 4, Write: 5},
		{BaseModel: model.BaseModel{CreatedAt: base.Add(time.Minute)}, Name: "sda", Read: 6, Write: 7},
	}
	if err := global.MonitorDB.CreateInBatches(ioRows, len(ioRows)).Error; err != nil {
		t.Fatalf("seed io: %v", err)
	}
	netRows := []model.MonitorNetwork{
		{BaseModel: model.BaseModel{CreatedAt: base}, Name: "all", Up: 1, Down: 1},
		{BaseModel: model.BaseModel{CreatedAt: base}, Name: "eth0", Up: 2, Down: 3},
		{BaseModel: model.BaseModel{CreatedAt: base}, Name: "veth-old", Up: 4, Down: 5},
	}
	if err := global.MonitorDB.CreateInBatches(netRows, len(netRows)).Error; err != nil {
		t.Fatalf("seed network: %v", err)
	}
	gpuRows := []model.MonitorGPU{
		{BaseModel: model.BaseModel{CreatedAt: base}, ProductName: "NVIDIA RTX 4090", GPUUtil: 10},
		{BaseModel: model.BaseModel{CreatedAt: base}, ProductName: "NVIDIA RTX 3090", GPUUtil: 20},
	}
	if err := global.GPUMonitorDB.CreateInBatches(gpuRows, len(gpuRows)).Error; err != nil {
		t.Fatalf("seed gpu: %v", err)
	}
	baseRows := []model.MonitorBase{
		{BaseModel: model.BaseModel{CreatedAt: base}, Cpu: 10, Memory: 20, LoadUsage: 30},
	}
	if err := global.MonitorDB.CreateInBatches(baseRows, len(baseRows)).Error; err != nil {
		t.Fatalf("seed base: %v", err)
	}
	return base
}

func searchRange(base time.Time) (time.Time, time.Time) {
	return base.Add(-time.Hour), base.Add(time.Hour)
}

func ioRowNames(t *testing.T, items []dto.MonitorData) map[string]int {
	t.Helper()
	counts := make(map[string]int)
	for _, item := range items {
		if item.Param != "io" {
			continue
		}
		for _, v := range item.Value {
			row := v.(model.MonitorIO)
			counts[row.Name]++
		}
	}
	return counts
}

func TestLoadMonitorDataIOWildcard(t *testing.T) {
	setupMonitorDBs(t)
	base := seedMonitorRows(t)
	start, end := searchRange(base)
	svc := NewIMonitorService()

	all, err := svc.LoadMonitorData(dto.MonitorSearch{Param: "io", IO: "", StartTime: start, EndTime: end})
	if err != nil {
		t.Fatalf("wildcard query: %v", err)
	}
	counts := ioRowNames(t, all)
	if counts["sda"] != 2 || counts["veth-old"] != 1 || counts["all"] != 1 {
		t.Fatalf("expected all stored devices in wildcard result, got %v", counts)
	}

	one, err := svc.LoadMonitorData(dto.MonitorSearch{Param: "io", IO: "sda", StartTime: start, EndTime: end})
	if err != nil {
		t.Fatalf("named query: %v", err)
	}
	if counts := ioRowNames(t, one); len(counts) != 1 || counts["sda"] != 2 {
		t.Fatalf("expected only sda rows, got %v", counts)
	}

	agg, err := svc.LoadMonitorData(dto.MonitorSearch{Param: "io", IO: "all", StartTime: start, EndTime: end})
	if err != nil {
		t.Fatalf("aggregate query: %v", err)
	}
	if counts := ioRowNames(t, agg); len(counts) != 1 || counts["all"] != 1 {
		t.Fatalf("expected only aggregate rows, got %v", counts)
	}
}

func TestLoadMonitorDataNetworkWildcard(t *testing.T) {
	setupMonitorDBs(t)
	base := seedMonitorRows(t)
	start, end := searchRange(base)
	svc := NewIMonitorService()

	items, err := svc.LoadMonitorData(dto.MonitorSearch{Param: "network", Network: "", StartTime: start, EndTime: end})
	if err != nil {
		t.Fatalf("wildcard query: %v", err)
	}
	counts := make(map[string]int)
	for _, item := range items {
		if item.Param != "network" {
			continue
		}
		for _, v := range item.Value {
			counts[v.(model.MonitorNetwork).Name]++
		}
	}
	if counts["eth0"] != 1 || counts["veth-old"] != 1 || counts["all"] != 1 {
		t.Fatalf("expected all stored interfaces, got %v", counts)
	}
}

func TestLoadMonitorDataMemoryParam(t *testing.T) {
	setupMonitorDBs(t)
	base := seedMonitorRows(t)
	start, end := searchRange(base)
	svc := NewIMonitorService()

	items, err := svc.LoadMonitorData(dto.MonitorSearch{Param: "memory", IO: "all", Network: "all", StartTime: start, EndTime: end})
	if err != nil {
		t.Fatalf("memory query: %v", err)
	}
	var baseItem *dto.MonitorData
	for i := range items {
		if items[i].Param == "base" {
			baseItem = &items[i]
		}
	}
	if baseItem == nil || len(baseItem.Value) != 1 {
		t.Fatalf("expected one base row for param=memory, got %+v", items)
	}
}

func TestMonitorOptionsIncludeRemovedDevices(t *testing.T) {
	setupMonitorDBs(t)
	seedMonitorRows(t)
	svc := NewIMonitorService()

	contains := func(list []string, want string) bool {
		for _, item := range list {
			if item == want {
				return true
			}
		}
		return false
	}
	ioOpts := svc.LoadIOOptions()
	if ioOpts[0] != "all" || !contains(ioOpts, "veth-old") || !contains(ioOpts, "sda") {
		t.Fatalf("io options should keep historical devices with 'all' first, got %v", ioOpts)
	}
	if !sort.StringsAreSorted(ioOpts[1:]) {
		t.Fatalf("io options after 'all' should be sorted, got %v", ioOpts)
	}
	netOpts := svc.LoadNetworkOptions()
	if netOpts[0] != "all" || !contains(netOpts, "veth-old") {
		t.Fatalf("network options should keep historical interfaces, got %v", netOpts)
	}
	gpuOpts := svc.LoadGPUOptions()
	if !contains(gpuOpts.Options, "NVIDIA RTX 4090") || !contains(gpuOpts.Options, "NVIDIA RTX 3090") {
		t.Fatalf("gpu options should keep historical products, got %v", gpuOpts.Options)
	}
}

func TestMonitorCompositeIndexesExist(t *testing.T) {
	setupMonitorDBs(t)
	seedMonitorRows(t)
	for _, name := range []string{"idx_monitor_ios_name_created", "idx_monitor_networks_name_created"} {
		var count int64
		if err := global.MonitorDB.Raw("SELECT COUNT(1) FROM sqlite_master WHERE type='index' AND name=?", name).Scan(&count).Error; err != nil {
			t.Fatalf("query sqlite_master: %v", err)
		}
		if count != 1 {
			t.Fatalf("index %s missing", name)
		}
	}
	var count int64
	if err := global.GPUMonitorDB.Raw("SELECT COUNT(1) FROM sqlite_master WHERE type='index' AND name='idx_monitor_gpus_product_created'").Scan(&count).Error; err != nil {
		t.Fatalf("query sqlite_master: %v", err)
	}
	if count != 1 {
		t.Fatal("index idx_monitor_gpus_product_created missing")
	}
}
