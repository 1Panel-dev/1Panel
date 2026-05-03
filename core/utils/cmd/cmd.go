package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

func SudoHandleCmd() string {
	cmd := exec.Command("sudo", "-n", "ls")
	if err := cmd.Run(); err == nil {
		return "sudo "
	}
	return ""
}

func Which(name string) bool {
	// Prefer Go's built-in PATH lookup so we don't depend on an external
	// `which` binary, which is not installed by default on minimal
	// distributions (e.g. Arch Linux, some Alpine images, slim containers).
	// See 1Panel-dev/1Panel#12605.
	if _, err := exec.LookPath(name); err == nil {
		return true
	}
	// Fall back to shelling out for environments where PATH inside the
	// agent process differs from the user's interactive shell PATH (the
	// previous behaviour, preserved for compatibility).
	stdout, err := RunDefaultWithStdoutBashCf("which %s", name)
	if err != nil || (len(strings.ReplaceAll(stdout, "\n", "")) == 0) {
		return false
	}
	return true
}

func ExecWithStreamOutput(command string, outputCallback func(string)) error {
	cmd := exec.Command("bash", "-c", command)

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
