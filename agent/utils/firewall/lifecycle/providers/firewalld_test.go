package providers

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestFirewalldStoppedRecognizesNormalInactiveResult(t *testing.T) {
	for _, test := range []struct {
		stdout string
		err    error
	}{
		{stdout: "not running\n"},
		{err: errors.New("stderr: FirewallD is not running, exit status 252")},
	} {
		if !firewalldStopped(test.stdout, test.err) {
			t.Fatalf("expected stopped result for stdout=%q err=%v", test.stdout, test.err)
		}
	}
}

func TestFirewalldStoppedKeepsUnexpectedFailures(t *testing.T) {
	if firewalldStopped("", errors.New("permission denied")) {
		t.Fatal("unexpected command failures must not be treated as an inactive firewall")
	}
}

func TestReplaceFirewalldConfigCreatesCleanConfigurationAndBackup(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "firewalld")
	backupDir := filepath.Join(root, "firewalld.backup")
	originalZone := filepath.Join(configDir, "zones", "custom.xml")
	if err := os.MkdirAll(filepath.Dir(originalZone), 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(originalZone, []byte("custom"), 0600); err != nil {
		t.Fatal(err)
	}

	prepared := false
	rollback, err := replaceFirewalldConfig(
		configDir,
		backupDir,
		func(path string) error {
			prepared = path == configDir
			return nil
		},
		func() error {
			for _, directory := range firewalldConfigSubdirectories {
				if info, err := os.Stat(filepath.Join(configDir, directory)); err != nil || !info.IsDir() {
					t.Fatalf("expected clean %s directory, info=%v err=%v", directory, info, err)
				}
			}
			if _, err := os.Stat(originalZone); !os.IsNotExist(err) {
				t.Fatalf("custom zone must not remain in clean configuration: %v", err)
			}
			return nil
		},
	)
	if err != nil {
		t.Fatalf("replace firewalld configuration: %v", err)
	}
	if !prepared {
		t.Fatal("expected clean configuration to be prepared")
	}
	if content, err := os.ReadFile(filepath.Join(backupDir, "zones", "custom.xml")); err != nil || string(content) != "custom" {
		t.Fatalf("expected original configuration in backup, content=%q err=%v", content, err)
	}

	if err := rollback(); err != nil {
		t.Fatalf("rollback firewalld configuration: %v", err)
	}
	if content, err := os.ReadFile(originalZone); err != nil || string(content) != "custom" {
		t.Fatalf("expected original configuration after rollback, content=%q err=%v", content, err)
	}
}

func TestReplaceFirewalldConfigRollsBackValidationFailure(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "firewalld")
	backupDir := filepath.Join(root, "firewalld.backup")
	originalConfig := filepath.Join(configDir, "firewalld.conf")
	if err := os.Mkdir(configDir, 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(originalConfig, []byte("DefaultZone=custom\n"), 0600); err != nil {
		t.Fatal(err)
	}

	rollback, err := replaceFirewalldConfig(
		configDir,
		backupDir,
		nil,
		func() error { return errors.New("invalid defaults") },
	)
	if err == nil || rollback != nil {
		t.Fatalf("expected validation failure with automatic rollback, hasRollback=%t err=%v", rollback != nil, err)
	}
	if content, readErr := os.ReadFile(originalConfig); readErr != nil || string(content) != "DefaultZone=custom\n" {
		t.Fatalf("expected original configuration after failed validation, content=%q err=%v", content, readErr)
	}
	if _, statErr := os.Stat(backupDir); !os.IsNotExist(statErr) {
		t.Fatalf("backup must be restored after failed validation: %v", statErr)
	}
}
