package dns

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
)

const (
	PdnsImage         = "powerdns/pdns-auth:latest"
	PdnsContainerName = "1panel-pdns"
	PdnsVolumeName    = "1panel-pdns-data"
	PdnsAPIPort       = "8081"
)

type ContainerStatus struct {
	Running bool
	Version string
}

func EnsurePdnsContainer(apiKey string) error {
	ctx := context.Background()
	cli, err := docker.NewDockerClient()
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}
	defer cli.Close()

	// Check if container already exists
	f := filters.NewArgs()
	f.Add("name", PdnsContainerName)
	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true, Filters: f})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	if len(containers) > 0 {
		c := containers[0]
		if c.State != "running" {
			if err := cli.ContainerStart(ctx, c.ID, container.StartOptions{}); err != nil {
				return fmt.Errorf("failed to start existing PowerDNS container: %w", err)
			}
		}
		return nil
	}

	// Pull image
	reader, err := cli.ImagePull(ctx, PdnsImage, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull PowerDNS image: %w", err)
	}
	_, _ = io.Copy(io.Discard, reader)
	_ = reader.Close()

	// Create container
	containerConfig := &container.Config{
		Image: PdnsImage,
		Env: []string{
			"PDNS_AUTH_API_KEY=" + apiKey,
		},
		Cmd: []string{
			"--api=yes",
			"--api-key=" + apiKey,
			"--webserver=yes",
			"--webserver-address=0.0.0.0",
			"--webserver-port=8081",
			"--webserver-allow-from=0.0.0.0/0",
			"--launch=gsqlite3",
			"--gsqlite3-database=/var/lib/powerdns/pdns.sqlite3",
			"--default-soa-edit=DEFAULT",
			"--default-soa-edit-signed=DEFAULT",
		},
		ExposedPorts: nat.PortSet{
			"53/tcp":   {},
			"53/udp":   {},
			"8081/tcp": {},
		},
	}

	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			"53/tcp": []nat.PortBinding{
				{HostIP: "0.0.0.0", HostPort: "53"},
			},
			"53/udp": []nat.PortBinding{
				{HostIP: "0.0.0.0", HostPort: "53"},
			},
			"8081/tcp": []nat.PortBinding{
				{HostIP: "127.0.0.1", HostPort: PdnsAPIPort},
			},
		},
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeVolume,
				Source: PdnsVolumeName,
				Target: "/var/lib/powerdns",
			},
		},
		RestartPolicy: container.RestartPolicy{
			Name: container.RestartPolicyMode("always"),
		},
	}

	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, &network.NetworkingConfig{}, nil, PdnsContainerName)
	if err != nil {
		return fmt.Errorf("failed to create PowerDNS container: %w", err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start PowerDNS container: %w", err)
	}

	// Wait briefly for the container to be ready
	time.Sleep(3 * time.Second)

	return nil
}

func StopPdnsContainer() error {
	ctx := context.Background()
	cli, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	timeout := 10
	return cli.ContainerStop(ctx, PdnsContainerName, container.StopOptions{Timeout: &timeout})
}

func RemovePdnsContainer() error {
	ctx := context.Background()
	cli, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	timeout := 10
	_ = cli.ContainerStop(ctx, PdnsContainerName, container.StopOptions{Timeout: &timeout})
	return cli.ContainerRemove(ctx, PdnsContainerName, container.RemoveOptions{Force: true})
}

func GetContainerStatus() ContainerStatus {
	ctx := context.Background()
	cli, err := docker.NewDockerClient()
	if err != nil {
		return ContainerStatus{}
	}
	defer cli.Close()

	f := filters.NewArgs()
	f.Add("name", PdnsContainerName)
	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true, Filters: f})
	if err != nil || len(containers) == 0 {
		return ContainerStatus{}
	}

	return ContainerStatus{
		Running: containers[0].State == "running",
		Version: containers[0].Image,
	}
}
