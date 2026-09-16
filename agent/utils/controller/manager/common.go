package manager

import (
	"errors"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

func handlerErr(out string, err error) error {
	if err != nil {
		if out != "" {
			return errors.New(out)
		}
		return err
	}
	return nil
}

func run(name string, args ...string) (string, error) {
	return runWithTimeout(10*time.Second, name, args...)
}

func runWithTimeout(timeout time.Duration, name string, args ...string) (string, error) {
	return cmd.NewCommandMgr(cmd.WithTimeout(timeout), cmd.WithEnv("LANGUAGE=en_US:en")).RunWithStdout(name, args...)
}

func serviceOperationTimeout(operation, serviceName string) time.Duration {
	if operation == "restart" {
		switch strings.TrimSuffix(serviceName, ".service") {
		case "docker", "dockerd", "docker.dockerd", "snap.docker.dockerd":
			return 2 * time.Minute
		}
	}
	return 10 * time.Second
}
