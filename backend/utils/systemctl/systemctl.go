package systemctl

import (
	"fmt"
	"github.com/pkg/errors"
	"os/exec"
	"strings"
)

func RunSystemCtl(args ...string) (string, error) {
	cmd := exec.Command("systemctl", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("failed to run command: %w", err)
	}
	return string(output), nil
}

func IsActive(serviceName string) (bool, error) {
	out, err := RunSystemCtl("is-active", serviceName)
	if err != nil {
		return false, err
	}
	return out == "active\n", nil
}

func IsExist(serviceName string) (bool, error) {
	out, err := RunSystemCtl("list-unit-files")
	if err != nil {
		return false, err
	}
	return strings.Contains(out, serviceName+".service"), nil
}

func handlerErr(out string, err error) error {
	if err != nil {
		if out != "" {
			return errors.New(out)
		}
		return err
	}
	return nil
}

func Restart(serviceName string) error {
	out, err := RunSystemCtl("restart", serviceName)
	return handlerErr(out, err)
}
func Start(serviceName string) error {
	out, err := RunSystemCtl("start", serviceName)
	return handlerErr(out, err)
}

func Stop(serviceName string) error {
	out, err := RunSystemCtl("stop", serviceName)
	return handlerErr(out, err)
}

func Operate(operate, serviceName string) error {
	out, err := RunSystemCtl(operate, serviceName)
	return handlerErr(out, err)
}
