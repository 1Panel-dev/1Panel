package forwarding

import (
	"errors"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/1Panel-dev/1Panel/agent/global"
)

var ErrRuleExists = errors.New("forwarding rule already exists")

type Operation struct {
	Rule
	Operation OperationType
}

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

func (m *Manager) Operate(operations []Operation, forceDelete bool) error {
	rules, err := m.adapter.List()
	if err != nil {
		return err
	}
	kept := rules[:0]
	for _, rule := range rules {
		shouldKeep := true
		for index := range operations {
			operation := &operations[index]
			if operation.TargetIP == "" {
				operation.TargetIP = "127.0.0.1"
			}
			if operation.Operation == OperationRemove && matches(*operation, rule) {
				shouldKeep = false
				break
			}
		}
		if shouldKeep {
			kept = append(kept, rule)
		}
	}
	for _, rule := range kept {
		for _, operation := range operations {
			if operation.Operation != OperationRemove && matches(operation, rule) {
				return ErrRuleExists
			}
		}
	}

	sort.SliceStable(operations, func(i, j int) bool {
		if operations[i].Operation == OperationRemove && operations[j].Operation != OperationRemove {
			return true
		}
		if operations[i].Operation != OperationRemove && operations[j].Operation == OperationRemove {
			return false
		}
		left, _ := strconv.Atoi(operations[i].Num)
		right, _ := strconv.Atoi(operations[j].Num)
		return left > right
	})
	for _, operation := range operations {
		for _, protocol := range strings.Split(operation.Protocol, "/") {
			rule := operation.Rule
			rule.Protocol = protocol
			if rule.TargetIP == "" {
				rule.TargetIP = "127.0.0.1"
			}
			if err := m.adapter.Operate(rule, operation.Operation); err != nil {
				if forceDelete && operation.Operation == OperationRemove {
					if global.LOG != nil {
						global.LOG.Error(err)
					}
					continue
				}
				return err
			}
		}
	}
	return nil
}

func (m *Manager) Enable() error { return m.adapter.Enable() }

func (m *Manager) Replay() error { return m.adapter.Replay() }

func (m *Manager) Name() string { return m.adapter.Name() }

func matches(operation Operation, rule Rule) bool {
	for _, protocol := range strings.Split(operation.Protocol, "/") {
		if operation.Port == rule.Port && operation.TargetPort == rule.TargetPort && operation.TargetIP == rule.TargetIP &&
			protocol == rule.Protocol && operation.Interface == rule.Interface {
			return true
		}
	}
	return false
}
