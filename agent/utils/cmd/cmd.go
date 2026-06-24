package cmd

import (
	"bytes"
	"fmt"
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

func commandStartDiagnostics(name, resolvedName, workDir string, env []string) string {
	items := []string{
		fmt.Sprintf("command=%q", name),
		fmt.Sprintf("resolved=%q", resolvedName),
		fmt.Sprintf("path=%q", envValue(env, "PATH")),
	}
	if workDir != "" {
		items = append(items, describeWorkDir(workDir))
	}
	return "; " + strings.Join(items, "; ")
}

func describeWorkDir(name string) string {
	statText := "stat=ok"
	if info, err := os.Stat(name); err != nil {
		statText = fmt.Sprintf("stat=%v", err)
	} else {
		statText = fmt.Sprintf("stat=mode:%s", info.Mode().String())
	}
	lstatText := "lstat=ok"
	if info, err := os.Lstat(name); err != nil {
		lstatText = fmt.Sprintf("lstat=%v", err)
	} else {
		lstatText = fmt.Sprintf("lstat=mode:%s", info.Mode().String())
	}
	return fmt.Sprintf("workdir=%q[%s %s]", name, statText, lstatText)
}

func envValue(env []string, key string) string {
	prefix := key + "="
	for i := len(env) - 1; i >= 0; i-- {
		if strings.HasPrefix(env[i], prefix) {
			return strings.TrimPrefix(env[i], prefix)
		}
	}
	return os.Getenv(key)
}
