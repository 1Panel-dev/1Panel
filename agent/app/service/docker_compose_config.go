package service

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
)

func composePortEnvKeys(index int) (containerPort, hostPort, hostIP, protocol string) {
	return fmt.Sprintf("CONTAINER_PORT_%d", index),
		fmt.Sprintf("HOST_PORT_%d", index),
		fmt.Sprintf("HOST_IP_%d", index),
		fmt.Sprintf("PORT_PROTOCOL_%d", index)
}

func isComposePortEnvKey(key string) bool {
	return strings.HasPrefix(key, "CONTAINER_PORT_") ||
		strings.HasPrefix(key, "HOST_PORT_") ||
		strings.HasPrefix(key, "HOST_IP_") ||
		strings.HasPrefix(key, "PORT_PROTOCOL_")
}

func formatComposePortMapping(hostIP, hostPort, containerPort, protocol string) string {
	return fmt.Sprintf("${%s}:${%s}:${%s}/%s", hostIP, hostPort, containerPort, normalizeComposeProtocol(protocol))
}

func normalizeComposeProtocol(protocol string) string {
	switch strings.ToLower(strings.TrimSpace(protocol)) {
	case "udp":
		return "udp"
	default:
		return "tcp"
	}
}

func formatComposeVolume(source, target, mode string) string {
	return fmt.Sprintf("%s:%s:%s", source, target, normalizeComposeVolumeMode(mode))
}

func normalizeComposeVolumeMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "ro":
		return "ro"
	default:
		return "rw"
	}
}

func loadComposeExposedPortsFromEnv(envs map[string]string, defaultHostIP string, strict bool) ([]request.ExposedPort, error) {
	var ports []request.ExposedPort
	for key, value := range envs {
		if !strings.HasPrefix(key, "CONTAINER_PORT_") {
			continue
		}
		index := strings.TrimPrefix(key, "CONTAINER_PORT_")
		containerPort, err := strconv.Atoi(value)
		if err != nil {
			if strict {
				return nil, err
			}
			continue
		}
		hostPort, err := strconv.Atoi(envs["HOST_PORT_"+index])
		if err != nil {
			if strict {
				return nil, err
			}
			continue
		}
		hostIP := envs["HOST_IP_"+index]
		if hostIP == "" {
			hostIP = defaultHostIP
		}
		ports = append(ports, request.ExposedPort{
			ContainerPort: containerPort,
			HostPort:      hostPort,
			HostIP:        hostIP,
			Protocol:      normalizeComposeProtocol(envs["PORT_PROTOCOL_"+index]),
		})
	}
	return ports, nil
}
