package mail

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
)

func execSetup(args ...string) (string, error) {
	ctx := context.Background()
	cli, err := docker.NewDockerClient()
	if err != nil {
		return "", fmt.Errorf("failed to create docker client: %w", err)
	}
	defer cli.Close()

	cmd := append([]string{"setup"}, args...)
	resp, err := cli.ContainerExecCreate(ctx, MailContainerName, container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return "", fmt.Errorf("exec create failed: %w", err)
	}

	hijack, err := cli.ContainerExecAttach(ctx, resp.ID, container.ExecAttachOptions{})
	if err != nil {
		return "", fmt.Errorf("exec attach failed: %w", err)
	}
	defer hijack.Close()

	raw, err := io.ReadAll(hijack.Reader)
	if err != nil {
		return "", err
	}
	var stdout, stderr bytes.Buffer
	_, _ = stdcopy.StdCopy(&stdout, &stderr, bytes.NewReader(raw))

	info, err := cli.ContainerExecInspect(ctx, resp.ID)
	if err != nil {
		return stdout.String(), err
	}
	if info.ExitCode != 0 {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = strings.TrimSpace(stdout.String())
		}
		if errMsg == "" {
			errMsg = fmt.Sprintf("command failed with exit code %d", info.ExitCode)
		}
		return "", fmt.Errorf("%s", errMsg)
	}
	return strings.TrimSpace(stdout.String()), nil
}

func AddAccount(email, password string) error {
	_, err := execSetup("email", "add", email, password)
	return err
}

func DeleteAccount(email string) error {
	_, err := execSetup("email", "del", "-y", email)
	return err
}

func UpdatePassword(email, password string) error {
	_, err := execSetup("email", "update", email, password)
	return err
}

func SetQuota(email string, quotaMB int) error {
	quota := fmt.Sprintf("%dM", quotaMB)
	if quotaMB == 0 {
		// Remove quota
		_, err := execSetup("quota", "del", email)
		return err
	}
	_, err := execSetup("quota", "set", email, quota)
	return err
}

func AddAlias(source, target string) error {
	_, err := execSetup("alias", "add", source, target)
	return err
}

func DeleteAlias(source, target string) error {
	_, err := execSetup("alias", "del", source, target)
	return err
}

func GenerateDKIM(domain string) error {
	_, err := execSetup("config", "dkim", "domain", domain)
	return err
}

func GetDKIMPublicKey(domain string) (string, error) {
	ctx := context.Background()
	cli, err := docker.NewDockerClient()
	if err != nil {
		return "", err
	}
	defer cli.Close()

	keyPath := fmt.Sprintf("/tmp/docker-mailserver/opendkim/keys/%s/mail.txt", domain)
	cmd := []string{"cat", keyPath}
	resp, err := cli.ContainerExecCreate(ctx, MailContainerName, container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return "", err
	}

	hijack, err := cli.ContainerExecAttach(ctx, resp.ID, container.ExecAttachOptions{})
	if err != nil {
		return "", err
	}
	defer hijack.Close()

	raw, _ := io.ReadAll(hijack.Reader)
	var stdout, stderr bytes.Buffer
	_, _ = stdcopy.StdCopy(&stdout, &stderr, bytes.NewReader(raw))

	return strings.TrimSpace(stdout.String()), nil
}
