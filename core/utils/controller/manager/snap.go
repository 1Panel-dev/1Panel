package manager

import (
	"strings"
)

type Snap struct{ toolCmd string }

func NewSnap() *Snap {
	return &Snap{toolCmd: "snap"}
}

func (s *Snap) IsExist(serviceName string) bool {
	out, err := run(s.toolCmd, "services")
	if err != nil {
		return false
	}
	return strings.Contains(out, serviceName)
}

func (s *Snap) IsActive(serviceName string) bool {
	out, err := run(s.toolCmd, "services")
	if err != nil {
		return false
	}
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if strings.Contains(line, serviceName) && strings.Contains(line, "active") {
			return true
		}
	}
	return false
}

func (s *Snap) IsEnable(serviceName string) bool {
	out, err := run(s.toolCmd, "services")
	if err != nil {
		return false
	}
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if strings.Contains(line, serviceName) && strings.Contains(line, "enabled") {
			return true
		}
	}
	return false
}

func (s *Snap) Operate(operate, serviceName string) error {
	if s.IsExist(serviceName) {
		return handlerErr(run(s.toolCmd, operate, serviceName))
	}
	return nil
}
