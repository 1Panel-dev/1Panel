package systemctl

import (
	"fmt"
	"github.com/pkg/errors"
	"os/exec"
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

func IsEnable(serviceName string) (bool, error) {
	out, err := RunSystemCtl("is-enabled", serviceName)
	if err != nil {
		return false, err
	}
	return out == "enabled\n", nil
}

// IsExist checks if a service exists.
func IsExist(serviceName string) (bool, error) {
	cmd := exec.Command("systemctl", "is-enabled", serviceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// If the command fails, check if the output indicates that the service does not exist.
		if string(output) == fmt.Sprintf("Failed to get unit file state for %s.service: No such file or directory\n", serviceName) {
			// Return false if the service does not exist.
			return false, nil
		}
		// Return an error if the command fails.
		return false, fmt.Errorf("failed to run command: %w", err)
	}
	// Return true if the service exists.
	return true, nil
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

func Operate(operate, serviceName string) error {
	out, err := RunSystemCtl(operate, serviceName)
	return handlerErr(out, err)
}
