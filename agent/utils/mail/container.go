package mail

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
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

const (
	MailServerImage  = "mailserver/docker-mailserver:latest"
	MailContainerName = "1panel-mailserver"
	MailVolumeName    = "1panel-mailserver-data"
	MailConfigVolume  = "1panel-mailserver-config"

	RoundcubeImage     = "roundcube/roundcubemail:latest"
	RoundcubeContainer = "1panel-roundcube"
	RoundcubePort      = "8443"

	MailNetworkName = "1panel-mail-network"
)

type ContainerStatus struct {
	MailRunning      bool
	MailVersion      string
	RoundcubeRunning bool
}

func EnsureMailContainers() error {
	ctx := context.Background()
	cli, err := docker.NewDockerClient()
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}
	defer cli.Close()

	// Ensure network
	networks, _ := cli.NetworkList(ctx, network.ListOptions{
		Filters: filters.NewArgs(filters.Arg("name", MailNetworkName)),
	})
	if len(networks) == 0 {
		_, err := cli.NetworkCreate(ctx, MailNetworkName, network.CreateOptions{Driver: "bridge"})
		if err != nil {
			return fmt.Errorf("failed to create mail network: %w", err)
		}
	}

	// Pull images
	for _, img := range []string{MailServerImage, RoundcubeImage} {
		reader, err := cli.ImagePull(ctx, img, image.PullOptions{})
		if err != nil {
			return fmt.Errorf("failed to pull %s: %w", img, err)
		}
		_, _ = io.Copy(io.Discard, reader)
		_ = reader.Close()
	}

	// Create/start mailserver
	if err := ensureOneContainer(ctx, cli, MailContainerName, &container.Config{
		Image:    MailServerImage,
		Hostname: MailContainerName,
		Env: []string{
			"ENABLE_RSPAMD=1",
			"ENABLE_CLAMAV=0",
			"ENABLE_FAIL2BAN=1",
			"ONE_DIR=1",
			"SSL_TYPE=",
			"PERMIT_DOCKER=network",
			"POSTMASTER_ADDRESS=postmaster@localhost",
		},
		ExposedPorts: nat.PortSet{
			"25/tcp": {}, "465/tcp": {}, "587/tcp": {}, "993/tcp": {},
		},
	}, &container.HostConfig{
		PortBindings: nat.PortMap{
			"25/tcp":  {{HostIP: "0.0.0.0", HostPort: "25"}},
			"465/tcp": {{HostIP: "0.0.0.0", HostPort: "465"}},
			"587/tcp": {{HostIP: "0.0.0.0", HostPort: "587"}},
			"993/tcp": {{HostIP: "0.0.0.0", HostPort: "993"}},
		},
		Mounts: []mount.Mount{
			{Type: mount.TypeVolume, Source: MailVolumeName, Target: "/var/mail-state"},
			{Type: mount.TypeVolume, Source: MailConfigVolume, Target: "/tmp/docker-mailserver"},
		},
		RestartPolicy: container.RestartPolicy{Name: container.RestartPolicyMode("always")},
	}, MailNetworkName); err != nil {
		return fmt.Errorf("failed to start mailserver: %w", err)
	}

	// Create/start roundcube
	if err := ensureOneContainer(ctx, cli, RoundcubeContainer, &container.Config{
		Image: RoundcubeImage,
		Env: []string{
			"ROUNDCUBEMAIL_DEFAULT_HOST=" + MailContainerName,
			"ROUNDCUBEMAIL_SMTP_SERVER=" + MailContainerName,
			"ROUNDCUBEMAIL_DEFAULT_PORT=993",
			"ROUNDCUBEMAIL_SMTP_PORT=587",
		},
		ExposedPorts: nat.PortSet{"80/tcp": {}},
	}, &container.HostConfig{
		PortBindings: nat.PortMap{
			"80/tcp": {{HostIP: "0.0.0.0", HostPort: RoundcubePort}},
		},
		RestartPolicy: container.RestartPolicy{Name: container.RestartPolicyMode("always")},
	}, MailNetworkName); err != nil {
		return fmt.Errorf("failed to start roundcube: %w", err)
	}

	time.Sleep(5 * time.Second)
	return nil
}

func ensureOneContainer(ctx context.Context, cli *client.Client, name string, config *container.Config, hostConfig *container.HostConfig, networkName string) error {
	f := filters.NewArgs(filters.Arg("name", name))
	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true, Filters: f})
	if err != nil {
		return err
	}
	if len(containers) > 0 {
		if containers[0].State != "running" {
			return cli.ContainerStart(ctx, containers[0].ID, container.StartOptions{})
		}
		return nil
	}

	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			networkName: {},
		},
	}
	resp, err := cli.ContainerCreate(ctx, config, hostConfig, networkConfig, nil, name)
	if err != nil {
		return err
	}
	return cli.ContainerStart(ctx, resp.ID, container.StartOptions{})
}

func StopMailContainers() error {
	ctx := context.Background()
	cli, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	defer cli.Close()
	timeout := 10
	_ = cli.ContainerStop(ctx, MailContainerName, container.StopOptions{Timeout: &timeout})
	_ = cli.ContainerStop(ctx, RoundcubeContainer, container.StopOptions{Timeout: &timeout})
	return nil
}

func RemoveMailContainers() error {
	ctx := context.Background()
	cli, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	defer cli.Close()
	timeout := 10
	_ = cli.ContainerStop(ctx, MailContainerName, container.StopOptions{Timeout: &timeout})
	_ = cli.ContainerStop(ctx, RoundcubeContainer, container.StopOptions{Timeout: &timeout})
	_ = cli.ContainerRemove(ctx, MailContainerName, container.RemoveOptions{Force: true})
	_ = cli.ContainerRemove(ctx, RoundcubeContainer, container.RemoveOptions{Force: true})
	return nil
}

func GetContainerStatus() ContainerStatus {
	ctx := context.Background()
	cli, err := docker.NewDockerClient()
	if err != nil {
		return ContainerStatus{}
	}
	defer cli.Close()

	status := ContainerStatus{}
	f := filters.NewArgs(filters.Arg("name", MailContainerName))
	containers, _ := cli.ContainerList(ctx, container.ListOptions{All: true, Filters: f})
	if len(containers) > 0 {
		status.MailRunning = containers[0].State == "running"
		status.MailVersion = containers[0].Image
	}

	f2 := filters.NewArgs(filters.Arg("name", RoundcubeContainer))
	containers2, _ := cli.ContainerList(ctx, container.ListOptions{All: true, Filters: f2})
	if len(containers2) > 0 {
		status.RoundcubeRunning = containers2[0].State == "running"
	}
	return status
}
