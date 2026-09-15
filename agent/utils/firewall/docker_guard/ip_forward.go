package docker_guard

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

const ipv4ForwardingPath = "/proc/sys/net/ipv4/ip_forward"

var ErrIPv4ForwardingDisabled = errors.New("IPv4 forwarding is disabled; set net.ipv4.ip_forward=1 before using Docker's firewall backend")

func CheckIPv4Forwarding() error {
	return checkIPv4Forwarding(os.ReadFile)
}

func checkIPv4Forwarding(readFile func(string) ([]byte, error)) error {
	value, err := readFile(ipv4ForwardingPath)
	if err != nil {
		return fmt.Errorf("inspect IPv4 forwarding: %w", err)
	}
	if strings.TrimSpace(string(value)) != "1" {
		return ErrIPv4ForwardingDisabled
	}
	return nil
}
