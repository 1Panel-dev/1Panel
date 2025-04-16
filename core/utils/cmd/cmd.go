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
