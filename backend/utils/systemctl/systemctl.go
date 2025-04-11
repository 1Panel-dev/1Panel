package systemctl

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/1Panel-dev/1Panel/backend/global"
)

func DefaultHandler(serviceName string) (*ServiceHandler, error) {
	svcName, err := smartServiceName(serviceName)
	if err != nil {
		// global.LOG.Errorf("SmartServiceName failed for %s: %v", serviceName, err)
		return nil, ErrServiceNotFound
	}
	return NewServiceHandler(defaultServiceConfig(svcName)), nil
}

func GetServiceName(serviceName string) (string, error) {
	serviceName, err := smartServiceName(serviceName)
	if err != nil {
		// global.LOG.Errorf("GetServiceName validation failed: %v", err)
		return "", ErrServiceNotFound
	}
	return serviceName, nil
}

func GetServicePath(serviceName string) (string, error) {
	handler, err := DefaultHandler(serviceName)
	if err != nil {
		// global.LOG.Errorf("GetServicePath handler init failed: %v", err)
		return "", ErrServiceNotFound
	}
	return handler.GetServicePath()
}

func CustomAction(action string, serviceName string) (ServiceResult, error) {
	handler, err := DefaultHandler(serviceName)
	if err != nil {
		global.LOG.Errorf("CustomAction handler init failed: %v", err)
		return ServiceResult{}, ErrServiceNotFound
	}
	result, err := handler.ExecuteAction(action)
	if err != nil {
		global.LOG.Errorf("CustomAction %s failed: %v", action, err)
		return result, fmt.Errorf("%s operation failed: %w | Output: %s", action, err, result.Output)
	}
	return result, nil
}

func IsExist(serviceName string) (bool, error) {
	handler, err := DefaultHandler(serviceName)
	if err != nil {
		return false, nil
	}
	result, _ := handler.IsExists()
	if result.IsExists {
		return true, nil
	} else {
		return false, nil
	}
}

func Start(serviceName string) error {
	handler, _ := DefaultHandler(serviceName)
	result, err := handler.StartService()
	if err != nil {
		global.LOG.Errorf("Service start failed: %v | Output: %s", err, result.Output)
		return fmt.Errorf("start failed: %v | Output: %s", err, result.Output)
	}
	return nil
}

func Stop(serviceName string) error {
	handler, err := DefaultHandler(serviceName)
	if err != nil {
		global.LOG.Errorf("Stop handler init failed: %v", err)
		return fmt.Errorf("%s is not exist", serviceName)
	}
	result, err := handler.StopService()
	if err != nil {
		global.LOG.Errorf("Service stop failed: %v", err)
		return fmt.Errorf("stop failed: %v | Output: %s", err, result.Output)
	}
	return nil
}

func Restart(serviceName string) error {
	handler, err := DefaultHandler(serviceName)
	if err != nil {
		global.LOG.Errorf("Restart handler init failed: %v", err)
		return fmt.Errorf("%s is not exist", serviceName)
	}
	result, err := handler.RestartService()
	if err != nil {
		global.LOG.Errorf("Service restart failed: %v", err)
		return fmt.Errorf("restart failed: %v | Output: %s", err, result.Output)
	}
	return nil
}

func SafeRestart(service string, configPaths []string) error {
	for _, path := range configPaths {
		if !FileExist(path) {
			global.LOG.Errorf("Config file missing: %s", path)
			return fmt.Errorf("config file missing: %s", path)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := executeCommand(ctx, "check", service); err != nil {
		global.LOG.Errorf("Config test failed: %v", err)
		return fmt.Errorf("config test failed: %w", err)
	}

	if err := Restart(service); err != nil {
		global.LOG.Errorf("SafeRestart failed: %v", err)
		return err
	}

	isActive, _, err := Status(service)
	if err != nil || !isActive {
		global.LOG.Error("Service not active after safe restart")
		return fmt.Errorf("service not active after restart")
	}

	return nil
}

func Enable(serviceName string) error {
	handler, err := DefaultHandler(serviceName)
	if err != nil {
		global.LOG.Errorf("Enable handler init failed: %v", err)
		return fmt.Errorf("%s is not exist", serviceName)
	}
	result, err := handler.EnableService()
	if err != nil {
		global.LOG.Errorf("Service enable failed: %v | Output: %s", err, result.Output)
		return fmt.Errorf("%s enable failed: %v ", serviceName, err)
	}
	return nil
}

func Disable(serviceName string) error {
	handler, _ := DefaultHandler(serviceName)
	result, err := handler.DisableService()
	if err != nil {
		global.LOG.Errorf("Service disable failed: %v", err)
		return fmt.Errorf("disable failed: %v | Output: %s", err, result.Output)
	}
	return nil
}

func Status(serviceName string) (isActive bool, isEnabled bool, err error) {
	handler, err := DefaultHandler(serviceName)
	if err != nil {
		global.LOG.Errorf("Status handler init failed: %v", err)
		return false, false, fmt.Errorf("%s is not exist", serviceName)
	}
	status, err := handler.CheckStatus()
	if err != nil {
		global.LOG.Errorf("Status check failed: %v", err)
		return false, false, fmt.Errorf("status check failed: %v | Output: %s", err, status.Output)
	}
	return status.IsActive, status.IsEnabled, nil
}

func IsActive(serviceName string) (bool, error) {
	handler, err := DefaultHandler(serviceName)
	if err != nil {
		return false, nil
	}
	status, err := handler.IsActive()
	if err != nil {
		return false, nil
	}
	return status.IsActive, nil
}

func IsEnable(serviceName string) (bool, error) {
	handler, err := DefaultHandler(serviceName)
	if err != nil {
		return false, nil
	}
	status, err := handler.IsEnabled()
	if err != nil {
		return false, nil
	}
	return status.IsEnabled, nil
}

type LogOption struct {
	TailLines string
}

func ViewLog(path string, opt LogOption) (string, error) {
	if !FileExist(path) {
		return "", fmt.Errorf("log file not found: %s", path)
	}
	args := []string{"-n", opt.TailLines, path}
	if opt.TailLines == "+1" {
		args = []string{"-n", "1", path}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	output, err := executeCommand(ctx, "tail", args...)
	if err != nil {
		return "", fmt.Errorf("tail failed: %w | Output: %s", err, string(output))
	}
	return string(output), nil
}

func FileExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
