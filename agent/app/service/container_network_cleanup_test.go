package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/errdefs"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type networkTaskClient struct{ fail bool }

func (f networkTaskClient) NetworkList(context.Context, network.ListOptions) ([]network.Inspect, error) {
	return []network.Inspect{{ID: "reserved", Name: "1panel-network", Scope: "local"}, {ID: "connected", Name: "busy", Scope: "local"}, {ID: "unused", Name: "free", Scope: "local"}, {ID: "race", Name: "race", Scope: "local"}}, nil
}
func (f networkTaskClient) ContainerList(context.Context, container.ListOptions) ([]container.Summary, error) {
	return nil, nil
}
func (f networkTaskClient) NetworkInspect(_ context.Context, id string, _ network.InspectOptions) (network.Inspect, error) {
	n := network.Inspect{}
	if id == "connected" {
		n.Containers = map[string]network.EndpointResource{"container-id": {}}
	}
	return n, nil
}
func (f networkTaskClient) NetworkRemove(_ context.Context, id string) error {
	if id == "race" {
		return errdefs.Conflict(errors.New("has active endpoints"))
	}
	if id != "unused" {
		return errors.New("unexpected removal")
	}
	if f.fail {
		return errors.New("remove failed")
	}
	return nil
}

func TestNetworkCleanupTaskPersistsLogsAndStatus(t *testing.T) {
	oldDB, oldTaskDB, oldDir, oldI18n := global.DB, global.TaskDB, global.Dir, global.I18n
	t.Cleanup(func() { global.DB = oldDB; global.TaskDB = oldTaskDB; global.Dir = oldDir; global.I18n = oldI18n })
	dir := t.TempDir()
	db, err := gorm.Open(sqlite.Open(filepath.Join(dir, "tasks.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	if err := db.AutoMigrate(&model.Task{}); err != nil {
		t.Fatal(err)
	}
	global.DB = nil
	global.TaskDB = db
	global.Dir.TaskDir = dir
	i18n.Init()
	for _, fail := range []bool{false, true} {
		name := "success"
		wantStatus := constant.StatusSuccess
		if fail {
			name = "partial failure"
			wantStatus = constant.StatusFailed
		}
		t.Run(name, func(t *testing.T) {
			item, err := task.NewTaskWithOps("Network", task.TaskClean, task.TaskScopeContainer, "", 0)
			if err != nil {
				t.Fatal(err)
			}
			item.AddSubTask("Clean", func(t *task.Task) error { return executeNetworkCleanup(t, networkTaskClient{fail: fail}) }, nil)
			err = item.Execute()
			if (err != nil) != fail {
				t.Fatalf("unexpected execution error: %v", err)
			}
			var saved model.Task
			if err := db.First(&saved, "id = ?", item.TaskID).Error; err != nil {
				t.Fatal(err)
			}
			if saved.Status != wantStatus {
				t.Fatalf("status %s, want %s", saved.Status, wantStatus)
			}
			content, err := os.ReadFile(saved.LogFile)
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range []string{"[busy] (connected): containers connected", "[race] (race): containers connected", "[1panel-network] (reserved): reserved network", "Network cleanup finished:", "[TASK-END]"} {
				if !strings.Contains(string(content), want) {
					t.Fatalf("missing %q in log: %s", want, content)
				}
			}
			want := "Deleted network [free]"
			if fail {
				want = "Failed to remove network [free]"
			}
			if !strings.Contains(string(content), want) {
				t.Fatalf("missing %q", want)
			}
		})
	}
}
