package docker_guard

import (
	"errors"
	"fmt"
	"strings"
)

var ErrDockerForwardPolicyDrop = errors.New("iptables FORWARD default policy is DROP")

func (m *NftablesManager) checkForwardPolicy() error {
	for _, family := range []struct{ command, name string }{
		{"iptables", FamilyIPv4},
		{"ip6tables", FamilyIPv6},
	} {
		if !m.runner.Exists(family.command) {
			continue
		}
		output, err := m.runner.Run(family.command, "-t", "filter", "-w", "-S", "FORWARD")
		if err != nil {
			return &FamilyError{Family: family.name, Err: fmt.Errorf("inspect iptables FORWARD policy: %w", err)}
		}
		found := false
		for _, line := range strings.Split(output, "\n") {
			fields := strings.Fields(line)
			if len(fields) != 3 || fields[0] != "-P" || fields[1] != "FORWARD" {
				continue
			}
			found = true
			if fields[2] == "DROP" {
				return &FamilyError{Family: family.name, Err: ErrDockerForwardPolicyDrop}
			}
			if fields[2] != "ACCEPT" {
				return &FamilyError{Family: family.name, Err: fmt.Errorf("unexpected iptables FORWARD policy: %s", fields[2])}
			}
		}
		if !found {
			return &FamilyError{Family: family.name, Err: errors.New("iptables FORWARD default policy was not found")}
		}
	}
	return nil
}
