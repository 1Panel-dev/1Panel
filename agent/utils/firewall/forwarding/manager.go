package forwarding

import (
	"errors"
	"strings"
	"sync"
)

var ErrRuleExists = errors.New("forwarding rule already exists")

type Status struct {
	Name     string
	Version  string
	IsActive bool
	IsInit   bool
	IsBind   bool
}

type RuntimeClient interface {
	Name() string
	Status() (bool, error)
	Version() (string, error)
}

type Manager struct {
	adapter Adapter
	runtime RuntimeClient
}

func NewManager(adapter Adapter, runtime RuntimeClient) *Manager {
	return &Manager{adapter: adapter, runtime: runtime}
}

func (m *Manager) Status() (Status, error) {
	status := Status{Name: m.adapter.Name(), Version: "-"}
	if m.runtime == nil {
		return status, nil
	}
	var versionErr error
	var statusErr error
	var initErr error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		status.Version, versionErr = m.runtime.Version()
	}()
	go func() {
		defer wg.Done()
		status.IsActive, statusErr = m.runtime.Status()
		status.IsInit, status.IsBind, initErr = m.adapter.InitStatus()
	}()
	wg.Wait()
	return status, errors.Join(versionErr, statusErr, initErr)
}

func (m *Manager) List(info, strategy string) ([]Rule, error) {
	rules, err := m.adapter.List()
	if err != nil {
		return nil, err
	}
	if strategy != "" {
		return []Rule{}, nil
	}
	filtered := make([]Rule, 0, len(rules))
	for _, rule := range rules {
		if info != "" && !strings.Contains(rule.Port, info) &&
			!strings.Contains(rule.TargetPort, info) && !strings.Contains(rule.TargetIP, info) {
			continue
		}
		filtered = append(filtered, rule)
	}
	return filtered, nil
}

func (m *Manager) Enable() error { return m.adapter.Enable() }

func (m *Manager) Reconcile(rules []Rule) error { return m.adapter.Reconcile(rules) }

func (m *Manager) Cleanup() error { return m.adapter.Cleanup() }

func (m *Manager) FamilyStatus(family string) (bool, bool, error) {
	return m.adapter.FamilyStatus(family)
}

func (m *Manager) Replay() error { return m.adapter.Replay() }

func (m *Manager) Name() string { return m.adapter.Name() }
