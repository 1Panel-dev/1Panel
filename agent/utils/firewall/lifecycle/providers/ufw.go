package providers

import (
	"fmt"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

type UFW struct {
	CmdStr string
}

func NewUFW() (*UFW, error) {
	var ufw UFW
	ufw.CmdStr = fmt.Sprintf("LANGUAGE=en_US:en %s ufw", cmd.SudoHandleCmd())
	return &ufw, nil
}

func (f *UFW) Name() string {
	return "ufw"
}

func (f *UFW) Status() (bool, error) {
	stdout, err := f.runWithStdout("status")
	if err != nil {
		return false, fmt.Errorf("load firewall status failed: %w", err)
	}
	if strings.Contains(stdout, "Status: active") {
		return true, nil
	}
	if strings.Contains(stdout, "状态： 激活") {
		return true, nil
	}
	return false, nil
}

func (f *UFW) Version() (string, error) {
	stdout, err := f.runWithStdout("version")
	if err != nil {
		return "", fmt.Errorf("load the firewall status failed, %v", err)
	}
	info := strings.Split(strings.TrimSpace(stdout), "\n")[0]
	return strings.ReplaceAll(info, "ufw ", ""), nil
}

func (f *UFW) Start() error {
	if err := f.run("--force", "enable"); err != nil {
		return fmt.Errorf("enable the firewall failed, %v", err)
	}
	return nil
}

func (f *UFW) Stop() error {
	if err := f.run("disable"); err != nil {
		return fmt.Errorf("stop the firewall failed, %v", err)
	}
	return nil
}

func (f *UFW) Reset() error {
	if err := f.run("--force", "reset"); err != nil {
		return fmt.Errorf("reset the firewall failed, %v", err)
	}
	return nil
}

func (f *UFW) Restart() error {
	if err := f.Stop(); err != nil {
		return err
	}
	if err := f.Start(); err != nil {
		return err
	}
	return nil
}

func (f *UFW) run(args ...string) error {
	_, err := f.runWithStdout(args...)
	return err
}

func (f *UFW) runWithStdout(args ...string) (string, error) {
	cmdMgr := cmd.NewCommandMgr(cmd.WithEnv("LANGUAGE=en_US:en"))
	return cmdMgr.RunWithOptionalSudoAndStdout("ufw", args...)
}
