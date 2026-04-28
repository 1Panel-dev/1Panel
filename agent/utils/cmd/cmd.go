package cmd

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
)

var (
	sudoHandleCmd string
	sudoCheckOnce sync.Once
)

func CheckIllegal(args ...string) bool {
	if args == nil {
		return false
	}
	for _, arg := range args {
		if strings.Contains(arg, "&") || strings.Contains(arg, "|") || strings.Contains(arg, ";") ||
			strings.Contains(arg, "$") || strings.Contains(arg, "'") || strings.Contains(arg, "`") ||
			strings.Contains(arg, "(") || strings.Contains(arg, ")") || strings.Contains(arg, "\"") ||
			strings.Contains(arg, "\n") || strings.Contains(arg, "\r") || strings.Contains(arg, ">") || strings.Contains(arg, "<") {
			return true
		}
	}
	return false
}

func SudoHandleCmd() string {
	sudoCheckOnce.Do(func() {
		cmd := exec.Command("sudo", "-n", "ls")
		if err := cmd.Run(); err == nil {
			sudoHandleCmd = "sudo "
		}
	})
	return sudoHandleCmd
}

func WrapWithOptionalSudo(name string, args ...string) (string, []string) {
	if SudoHandleCmd() == "" {
		return name, args
	}
	return "sudo", append([]string{"-n", name}, args...)
}

func ExecCommandWithOptionalSudo(name string, args ...string) *exec.Cmd {
	commandName, commandArgs := WrapWithOptionalSudo(name, args...)
	return exec.Command(commandName, commandArgs...)
}

func WriteFileWithOptionalSudo(name string, data []byte, perm os.FileMode) error {
	if err := os.WriteFile(name, data, perm); err == nil {
		return nil
	} else if SudoHandleCmd() == "" {
		return err
	}
	command := exec.Command("sudo", "-n", "tee", name)
	command.Stdin = bytes.NewReader(data)
	command.Stdout = io.Discard
	return command.Run()
}

func Which(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
