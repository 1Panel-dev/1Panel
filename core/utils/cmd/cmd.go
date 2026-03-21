package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
)

var (
	bashPathOnce  sync.Once
	bashPathCache string
)

// BashPath returns the full path to bash, falling back to common locations
// if "bash" is not found in $PATH. This fixes issues on minimal OS installs
// (e.g., AlmaLinux) where bash may not be in the default PATH.
func BashPath() string {
	bashPathOnce.Do(func() {
		if p, err := exec.LookPath("bash"); err == nil {
			bashPathCache = p
			return
		}
		for _, candidate := range []string{"/bin/bash", "/usr/bin/bash", "/usr/local/bin/bash"} {
			if _, err := os.Stat(candidate); err == nil {
				bashPathCache = candidate
				return
			}
		}
		bashPathCache = "bash"
	})
	return bashPathCache
}

func SudoHandleCmd() string {
	cmd := exec.Command("sudo", "-n", "ls")
	if err := cmd.Run(); err == nil {
		return "sudo "
	}
	return ""
}

func Which(name string) bool {
	stdout, err := RunDefaultWithStdoutBashCf("which %s", name)
	if err != nil || (len(strings.ReplaceAll(stdout, "\n", "")) == 0) {
		return false
	}
	return true
}

func ExecWithStreamOutput(command string, outputCallback func(string)) error {
	cmd := exec.Command(BashPath(), "-c", command)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	go streamReader(stdout, outputCallback)
	go streamReader(stderr, outputCallback)

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("command finished with error: %w", err)
	}
	return nil
}

func streamReader(reader io.ReadCloser, callback func(string)) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		callback(scanner.Text())
	}
}
