package systemctl

import (
	"fmt"
	"os"

	"github.com/1Panel-dev/1Panel/backend/global"
)

func DefaultHandler(serviceName string) (*ServiceHandler, error) {
	svcName, err := smartServiceName(serviceName)
	if err != nil {
		return nil, ErrServiceNotFound
	}
	return NewServiceHandler(defaultServiceConfig(svcName)), nil
}

func GetServiceName(serviceName string) (string, error) {
	serviceName, err := smartServiceName(serviceName)
	if err != nil {
		return "", ErrServiceNotFound
	}
	return serviceName, nil
}

func GetServicePath(serviceName string) (string, error) {
	handler, err := DefaultHandler(serviceName)
	if err != nil {
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

func FileExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
