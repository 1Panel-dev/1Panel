package providers

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/constant"
)

func ensureForwardingSysctls(system forwardingSystem, withIPv6 bool) error {
	paths := []string{"/proc/sys/net/ipv4/ip_forward"}
	if withIPv6 {
		paths = append(paths, "/proc/sys/net/ipv6/conf/all/forwarding")
	}
	for _, path := range paths {
		if err := system.WriteFile(path, []byte("1"), constant.FilePerm); err != nil {
			return fmt.Errorf("failed to enable IP forwarding at %s: %w", path, err)
		}
	}
	data, err := system.ReadFile("/etc/sysctl.conf")
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to read /etc/sysctl.conf: %w", err)
	}
	content := enableForwardingSysctls(string(data), withIPv6)
	if err := system.WriteFile("/etc/sysctl.conf", []byte(content), constant.FilePerm); err != nil {
		return fmt.Errorf("failed to persist IP forwarding: %w", err)
	}
	if err := system.RunWithOptionalSudo("sysctl", "-p"); err != nil {
		return fmt.Errorf("failed to apply IP forwarding: %w", err)
	}
	return nil
}

func enableForwardingSysctls(content string, withIPv6 bool) string {
	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	wanted := map[string]string{"net.ipv4.ip_forward": "net.ipv4.ip_forward = 1"}
	if withIPv6 {
		wanted["net.ipv6.conf.all.forwarding"] = "net.ipv6.conf.all.forwarding = 1"
	}
	found := make(map[string]bool, len(wanted))
	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		if replacement, ok := wanted[key]; ok {
			lines[index] = replacement
			found[key] = true
		}
	}
	for _, key := range []string{"net.ipv4.ip_forward", "net.ipv6.conf.all.forwarding"} {
		if replacement, ok := wanted[key]; ok && !found[key] {
			lines = append(lines, replacement)
		}
	}
	if len(lines) > 0 && lines[0] == "" {
		lines = lines[1:]
	}
	return strings.Join(lines, "\n") + "\n"
}
