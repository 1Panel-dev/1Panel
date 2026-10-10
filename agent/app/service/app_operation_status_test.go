package service

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

func operationStatusTestDB(t *testing.T) (*gorm.DB, *gorm.DB, model.AppInstall) {
	t.Helper()
	open := func(name string, schema interface{}) *gorm.DB {
		db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), name)), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := db.AutoMigrate(schema); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			sqlDB, _ := db.DB()
			if sqlDB != nil {
				_ = sqlDB.Close()
			}
		})
		return db
	}
	appDB := open("apps.db", &model.AppInstall{})
	taskDB := open("tasks.db", &model.Task{})
	install := model.AppInstall{
		Name: "gb10-ds-vision", AppId: 1, AppDetailId: 1,
		Version: "nvidia-gb10-dspark-0.1.1", ContainerName: "vllm-dspark-1", ServiceName: "vllm-dspark",
		Status: constant.StatusStarting, Env: `{"MODEL_DIR":"/models/vision"}`,
	}
	if err := appDB.Omit(clause.Associations).Create(&install).Error; err != nil {
		t.Fatal(err)
	}
	return appDB, taskDB, install
}

func TestAppInstallFailedTaskRecoversPendingState(t *testing.T) {
	for _, state := range []string{constant.StatusStarting, constant.StatusWaiting, constant.StatusRestarting, constant.StatusInstalling} {
		t.Run(state, func(t *testing.T) {
			appDB, taskDB, install := operationStatusTestDB(t)
			install.Status = state
			if err := appDB.Omit(clause.Associations).Save(&install).Error; err != nil {
				t.Fatal(err)
			}
			task := model.Task{ID: "failed", Type: "App", Operate: "TaskUpdate", ResourceID: install.ID,
				Status: constant.StatusFailed, ErrorMsg: "model startup script failed", CreatedAt: install.UpdatedAt.Add(time.Second), EndAt: install.UpdatedAt.Add(2 * time.Second)}
			if err := taskDB.Create(&task).Error; err != nil {
				t.Fatal(err)
			}
			if err := reconcileAppInstallTaskFailure(&install, task.Operate, appDB, taskDB); err != nil {
				t.Fatal(err)
			}
			want := constant.StatusUpErr
			if state == constant.StatusInstalling {
				want = constant.StatusInstallErr
			}
			var stored model.AppInstall
			if err := appDB.First(&stored, install.ID).Error; err != nil {
				t.Fatal(err)
			}
			if install.Status != want || stored.Status != want || stored.Message != task.ErrorMsg || stored.Env != install.Env {
				t.Fatalf("recovered state = %s/%s, message %q, env %q", install.Status, stored.Status, stored.Message, stored.Env)
			}
		})
	}
}

func TestAppInstallPendingStateKeepsActiveAndUnrelatedTasks(t *testing.T) {
	for _, taskStatus := range []string{constant.StatusExecuting, constant.StatusSuccess, "missing", "previous-failure", "other-resource"} {
		t.Run(taskStatus, func(t *testing.T) {
			appDB, taskDB, install := operationStatusTestDB(t)
			task := model.Task{ID: "task", Type: "App", Operate: "TaskUpdate", ResourceID: install.ID,
				Status: taskStatus, CreatedAt: install.UpdatedAt.Add(time.Second), EndAt: install.UpdatedAt.Add(2 * time.Second)}
			if taskStatus == "previous-failure" {
				task.Status = constant.StatusFailed
				task.CreatedAt = install.UpdatedAt.Add(-2 * time.Hour)
				task.EndAt = install.UpdatedAt.Add(-time.Hour)
			}
			if taskStatus == "other-resource" {
				task.Status = constant.StatusFailed
				task.ResourceID++
			}
			if taskStatus != "missing" {
				if err := taskDB.Create(&task).Error; err != nil {
					t.Fatal(err)
				}
			}
			if err := reconcileAppInstallTaskFailure(&install, task.Operate, appDB, taskDB); err != nil {
				t.Fatal(err)
			}
			if install.Status != constant.StatusStarting {
				t.Fatalf("unrelated or active task changed pending state to %s", install.Status)
			}
		})
	}
}

func TestAppInstallRecoveryDoesNotOverwriteConcurrentRetry(t *testing.T) {
	appDB, taskDB, install := operationStatusTestDB(t)
	task := model.Task{ID: "failed", Type: "App", Operate: "TaskUpdate", ResourceID: install.ID,
		Status: constant.StatusCanceled, ErrorMsg: "interrupted", CreatedAt: install.UpdatedAt.Add(time.Second), EndAt: install.UpdatedAt.Add(2 * time.Second)}
	if err := taskDB.Create(&task).Error; err != nil {
		t.Fatal(err)
	}
	if err := appDB.Callback().Update().Before("gorm:update").Register("new-operation", func(db *gorm.DB) {
		if err := db.Exec("UPDATE app_installs SET updated_at = ?, message = ? WHERE id = ?", install.UpdatedAt.Add(time.Hour), "new attempt", install.ID).Error; err != nil {
			t.Error(err)
		}
	}); err != nil {
		t.Fatal(err)
	}
	if err := reconcileAppInstallTaskFailure(&install, task.Operate, appDB, taskDB); err != nil {
		t.Fatal(err)
	}
	var stored model.AppInstall
	if err := appDB.First(&stored, install.ID).Error; err != nil {
		t.Fatal(err)
	}
	if stored.Status != constant.StatusStarting || stored.Message != "new attempt" {
		t.Fatalf("recovery overwrote a newer operation: %s, %s", stored.Status, stored.Message)
	}
}

func TestAppInstallRestartRecoveryCoversLifecycleStates(t *testing.T) {
	for _, state := range []string{constant.StatusStarting, constant.StatusWaiting, constant.StatusInstalling} {
		if !appInstallOperationInterruptedOnRestart(state) {
			t.Errorf("restart recovery omitted %s", state)
		}
	}
	for _, state := range []string{constant.StatusRunning, constant.StatusStopped, constant.StatusRestarting} {
		if appInstallOperationInterruptedOnRestart(state) {
			t.Errorf("restart recovery changes a Docker runtime state: %s", state)
		}
	}
}
