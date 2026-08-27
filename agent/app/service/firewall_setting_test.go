package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/docker_guard"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestFirewallSettingServiceKeepsSelectionsWhenBackendsAreMissing(t *testing.T) {
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
	for key, backend := range map[string]string{
		constant.FirewallSystemBackendKey:     constant.FirewallProviderNftables,
		constant.FirewallForwardingBackendKey: constant.FirewallProviderNftables,
		constant.FirewallDockerBackendKey:     constant.FirewallProviderNftables,
	} {
		if err := settingRepo.UpdateOrCreate(key, backend); err != nil {
			t.Fatalf("save selected backend %s: %v", key, err)
		}
	}

	settings, err := (&FirewallSettingService{}).Load(context.Background())
	if err != nil {
		t.Fatalf("load firewall settings: %v", err)
	}
	for subsystem, selected := range map[string]string{
		"system":     settings.System.Selected,
		"forwarding": settings.Forwarding.Selected,
		"docker":     settings.Docker.Selected,
	} {
		if selected != constant.FirewallProviderNftables {
			t.Fatalf("%s selected backend = %q, want persisted nftables", subsystem, selected)
		}
	}
	for _, option := range settings.Docker.Options {
		if option.Installed || option.Active || !option.Supported {
			t.Fatalf("unexpected Docker backend availability without firewall commands: %#v", option)
		}
	}
}

func TestFirewallSettingServiceSeparatesDockerAndFirewallAvailability(t *testing.T) {
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
	for _, name := range []string{"iptables", "iptables-restore"} {
		if err := os.WriteFile(filepath.Join(binDir, name), []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
			t.Fatalf("create fake %s executable: %v", name, err)
		}
	}
	t.Setenv("PATH", binDir)

	settings, err := (&FirewallSettingService{}).Load(context.Background())
	if err != nil {
		t.Fatalf("load firewall settings: %v", err)
	}
	for _, option := range settings.Docker.Options {
		if option.Name != constant.FirewallProviderIptables {
			continue
		}
		if !option.Installed || option.Supported || option.Active {
			t.Fatalf("Docker and firewall availability were not separated: %#v", option)
		}
		return
	}
	t.Fatal("iptables Docker backend option was not returned")
}

func TestFirewallSettingServiceUsesExplicitEmptyDockerAndIptablesForwardingDefaults(t *testing.T) {
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
	t.Setenv("PATH", t.TempDir())

	settings, err := (&FirewallSettingService{}).Load(context.Background())
	if err != nil {
		t.Fatalf("load firewall settings: %v", err)
	}
	if settings.Docker.Selected != "" {
		t.Fatalf("Docker selected backend = %q, want empty", settings.Docker.Selected)
	}
	if settings.Forwarding.Selected != constant.FirewallProviderIptables {
		t.Fatalf("forwarding selected backend = %q, want iptables", settings.Forwarding.Selected)
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
