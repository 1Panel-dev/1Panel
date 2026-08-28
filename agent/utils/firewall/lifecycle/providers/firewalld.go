package providers

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/controller"
)

type Firewalld struct{}

const firewalldConfigDir = "/etc/firewalld"

var firewalldConfigSubdirectories = []string{
	"helpers",
	"icmptypes",
	"ipsets",
	"policies",
	"services",
	"zones",
}

func NewFirewalld() (*Firewalld, error) {
	return &Firewalld{}, nil
}

func (f *Firewalld) Name() string {
	return "firewalld"
}

func (f *Firewalld) Status() (bool, error) {
	stdout, err := cmd.NewCommandMgr(cmd.WithEnv("LANGUAGE=en_US:en")).RunWithStdout("firewall-cmd", "--state")
	if err != nil {
		if firewalldStopped(stdout, err) {
			return false, nil
		}
		return false, fmt.Errorf("load firewall status failed: %w", err)
	}
	return strings.TrimSpace(stdout) == "running", nil
}

func firewalldStopped(stdout string, err error) bool {
	message := strings.ToLower(strings.TrimSpace(stdout))
	if err != nil {
		message += " " + strings.ToLower(err.Error())
	}
	return strings.Contains(message, "not running")
}

func (f *Firewalld) Version() (string, error) {
	stdout, err := cmd.NewCommandMgr(cmd.WithEnv("LANGUAGE=en_US:en")).RunWithStdout("firewall-cmd", "--version")
	if err != nil {
		return "", fmt.Errorf("load the firewall version failed, %v", err)
	}
	return strings.ReplaceAll(stdout, "\n ", ""), nil
}

func (f *Firewalld) Start() error {
	if err := controller.HandleStart("firewalld"); err != nil {
		return fmt.Errorf("enable the firewall failed, err: %v", err)
	}
	return nil
}

func (f *Firewalld) Stop() error {
	if err := controller.HandleStop("firewalld"); err != nil {
		return fmt.Errorf("stop the firewall failed, err: %v", err)
	}
	return nil
}

func (f *Firewalld) Reset() error {
	running, err := f.Status()
	if err != nil {
		return fmt.Errorf("reset firewalld to defaults failed: %w", err)
	}
	if running {
		if err := f.Stop(); err != nil {
			return fmt.Errorf("reset firewalld to defaults failed: %w", err)
		}
	}

	backupDir := fmt.Sprintf("%s.1panel-backup-%d", firewalldConfigDir, time.Now().UnixNano())
	rollback, err := replaceFirewalldConfig(
		firewalldConfigDir,
		backupDir,
		restoreFirewalldConfigContext,
		validateFirewalldConfig,
	)
	if err != nil {
		return fmt.Errorf("reset firewalld to defaults failed: %w", err)
	}
	if err := controller.Handle("disable", "firewalld"); err != nil {
		return fmt.Errorf(
			"reset firewalld to defaults failed: %w",
			errors.Join(err, rollback()),
		)
	}
	return nil
}

func replaceFirewalldConfig(
	configDir string,
	backupDir string,
	prepare func(string) error,
	validate func() error,
) (func() error, error) {
	info, err := os.Lstat(configDir)
	hadConfig := err == nil
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("inspect firewalld configuration: %w", err)
	}
	if hadConfig && !info.IsDir() {
		return nil, fmt.Errorf("firewalld configuration path %s is not a directory", configDir)
	}
	if hadConfig {
		if _, err := os.Lstat(backupDir); err == nil {
			return nil, fmt.Errorf("firewalld backup path already exists: %s", backupDir)
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("inspect firewalld backup path: %w", err)
		}
		if err := os.Rename(configDir, backupDir); err != nil {
			return nil, fmt.Errorf("back up firewalld configuration: %w", err)
		}
	}

	rollback := func() error {
		removeErr := os.RemoveAll(configDir)
		if !hadConfig {
			return removeErr
		}
		restoreErr := os.Rename(backupDir, configDir)
		return errors.Join(removeErr, restoreErr)
	}
	fail := func(cause error) (func() error, error) {
		if rollbackErr := rollback(); rollbackErr != nil {
			return nil, errors.Join(cause, fmt.Errorf("rollback firewalld configuration: %w", rollbackErr))
		}
		return nil, cause
	}

	mode := os.FileMode(0755)
	if hadConfig {
		mode = info.Mode().Perm()
	}
	if err := os.Mkdir(configDir, mode); err != nil {
		return fail(fmt.Errorf("create firewalld configuration directory: %w", err))
	}
	for _, directory := range firewalldConfigSubdirectories {
		if err := os.Mkdir(filepath.Join(configDir, directory), 0755); err != nil {
			return fail(fmt.Errorf("create firewalld %s directory: %w", directory, err))
		}
	}
	if prepare != nil {
		if err := prepare(configDir); err != nil {
			return fail(fmt.Errorf("prepare firewalld configuration directory: %w", err))
		}
	}
	if validate != nil {
		if err := validate(); err != nil {
			return fail(fmt.Errorf("validate default firewalld configuration: %w", err))
		}
	}
	return rollback, nil
}

func restoreFirewalldConfigContext(configDir string) error {
	if _, err := exec.LookPath("restorecon"); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return nil
		}
		return err
	}
	_, err := cmd.NewCommandMgr(cmd.WithEnv("LANGUAGE=en_US:en")).RunWithOptionalSudoAndStdout(
		"restorecon", "-RF", configDir,
	)
	return err
}

func validateFirewalldConfig() error {
	_, err := cmd.NewCommandMgr(cmd.WithEnv("LANGUAGE=en_US:en")).RunWithOptionalSudoAndStdout(
		"firewall-offline-cmd", "--get-zones",
	)
	return err
}

func (f *Firewalld) Restart() error {
	if err := controller.HandleRestart("firewalld"); err != nil {
		return fmt.Errorf("restart the firewall failed, err: %v", err)
	}
	return nil
}
