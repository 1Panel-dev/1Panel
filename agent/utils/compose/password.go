package compose

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

func RecreatePasswordService(ctx context.Context, filePath, projectName, service, dockerHost string, environment map[string]string) error {
	command, err := passwordServiceCommand(ctx, filePath, projectName, dockerHost, filepath.Join(filepath.Dir(filePath), ".env"), environment,
		"up", "-d", "--no-deps", "--no-build", "--pull", "never", "--force-recreate", "--timeout", "300", service)
	if err != nil {
		return err
	}
	return command.Run()
}

func ValidatePasswordService(ctx context.Context, filePath, projectName, service, dockerHost string, environment map[string]string, content []byte) error {
	candidate, err := os.CreateTemp("", "1panel-cluster-password-env-*")
	if err != nil {
		return err
	}
	candidatePath := candidate.Name()
	defer os.Remove(candidatePath)
	_, writeErr := candidate.Write(content)
	if err := errors.Join(writeErr, candidate.Close()); err != nil {
		return err
	}

	command, err := passwordServiceCommand(ctx, filePath, projectName, dockerHost, candidatePath, environment, "config", "--quiet", service)
	if err != nil {
		return err
	}
	return command.Run()
}

func passwordServiceCommand(ctx context.Context, filePath, projectName, dockerHost, envPath string, environment map[string]string, operation ...string) (*exec.Cmd, error) {
	if err := checkCmd(); err != nil {
		return nil, err
	}
	base, extra := getComposeBaseCmd()
	args := appendProjectName(append([]string(nil), extra...), []string{projectName})
	args = append(args, "--env-file", envPath, "-f", filePath)
	args = append(args, operation...)
	command := exec.CommandContext(ctx, base, args...)
	command.Env = slices.DeleteFunc(command.Environ(), func(value string) bool {
		key, _, _ := strings.Cut(value, "=")
		_, configured := environment[key]
		return configured
	})
	command.Env = append(command.Env, "DOCKER_HOST="+dockerHost, "DOCKER_CONTEXT=", "COMPOSE_REMOVE_ORPHANS=false")
	return command, nil
}
