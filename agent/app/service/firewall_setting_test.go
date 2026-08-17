package service

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/constant"
)

func withDockerFirewallTestHooks(t *testing.T, initial string) string {
	t.Helper()
	originalPath := constant.DaemonJsonPath
	originalValidate := validateDockerFirewallConfig
	originalRestart := restartDockerFirewall
	path := filepath.Join(t.TempDir(), "daemon.json")
	if err := os.WriteFile(path, []byte(initial), 0640); err != nil {
		t.Fatal(err)
	}
	constant.DaemonJsonPath = path
	validateDockerFirewallConfig = func() error { return nil }
	restartDockerFirewall = func(string) error { return nil }
	t.Cleanup(func() {
		constant.DaemonJsonPath = originalPath
		validateDockerFirewallConfig = originalValidate
		restartDockerFirewall = originalRestart
	})
	return path
}

func TestUpdateDockerFirewallBackendPreservesDaemonConfiguration(t *testing.T) {
	path := withDockerFirewallTestHooks(t, `{"live-restore":true}`)
	if err := updateDockerFirewallBackend("nftables", false); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, `"live-restore": true`) || !strings.Contains(content, `"firewall-backend": "nftables"`) {
		t.Fatalf("daemon configuration was not preserved: %s", content)
	}
}

func TestUpdateDockerFirewallBackendRestoresInvalidConfiguration(t *testing.T) {
	original := `{"live-restore":true}`
	path := withDockerFirewallTestHooks(t, original)
	validateDockerFirewallConfig = func() error { return errors.New("invalid Docker configuration") }
	if err := updateDockerFirewallBackend("nftables", false); err == nil {
		t.Fatal("expected validation failure")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != original {
		t.Fatalf("daemon configuration was not restored: %s", data)
	}
}

func TestUpdateDockerFirewallBackendRestoresAfterRestartFailure(t *testing.T) {
	original := `{"iptables":true}`
	path := withDockerFirewallTestHooks(t, original)
	restarts := 0
	restartDockerFirewall = func(string) error {
		restarts++
		if restarts == 1 {
			return errors.New("restart failed")
		}
		return nil
	}
	if err := updateDockerFirewallBackend("nftables", true); err == nil {
		t.Fatal("expected restart failure")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != original || restarts != 2 {
		t.Fatalf("restart rollback failed: config=%s restarts=%d", data, restarts)
	}
}
