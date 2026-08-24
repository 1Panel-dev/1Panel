package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/docker_guard"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestFirewallSettingServiceDockerOptionsRequireBackendCommands(t *testing.T) {
	previousDB := global.DB
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", uuid.NewString())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open settings database: %v", err)
	}
	if err := db.AutoMigrate(&model.Setting{}); err != nil {
		t.Fatalf("migrate settings database: %v", err)
	}
	global.DB = db
	t.Cleanup(func() { global.DB = previousDB })

	binDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(binDir, "docker"), []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("create fake Docker executable: %v", err)
	}
	t.Setenv("PATH", binDir)

	settings, err := (&FirewallSettingService{}).Load(context.Background())
	if err != nil {
		t.Fatalf("load firewall settings: %v", err)
	}
	for _, option := range settings.Docker.Options {
		if option.Installed {
			t.Fatalf("Docker backend %q reported installed without its firewall commands", option.Name)
		}
	}
}

func TestFirewallSettingServiceSelectsGuardBackendWithoutDockerRestart(t *testing.T) {
	previousDB := global.DB
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", uuid.NewString())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open settings database: %v", err)
	}
	if err := db.AutoMigrate(&model.Setting{}); err != nil {
		t.Fatalf("migrate settings database: %v", err)
	}
	global.DB = db
	t.Cleanup(func() { global.DB = previousDB })

	err = (&FirewallSettingService{}).Operate(context.Background(), dto.FirewallBackendOperation{
		Subsystem: "docker",
		Backend:   "nftables",
		Operation: "select",
	})
	if err != nil {
		t.Fatalf("select Docker guard backend: %v", err)
	}
	if got := selectedDockerFirewallBackend("iptables"); got != "nftables" {
		t.Fatalf("selected Docker guard backend = %q, want nftables", got)
	}
	if _, ok := (&DockerPortGuardService{}).guardRuntime(selectedDockerFirewallBackend("iptables")).(*docker_guard.NftablesManager); !ok {
		t.Fatal("Docker port guard did not use the selected nftables runtime")
	}
}

func TestFirewallSettingServiceRejectsServiceBackendInitialization(t *testing.T) {
	for _, backend := range []string{"firewalld", "ufw"} {
		for _, operation := range []string{"initialize", "cleanup"} {
			err := (&FirewallSettingService{}).Operate(context.Background(), dto.FirewallBackendOperation{
				Subsystem: "system",
				Backend:   backend,
				Operation: operation,
			})
			if err == nil {
				t.Fatalf("%s %s unexpectedly succeeded", backend, operation)
			}
		}
	}
}
