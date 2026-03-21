package cmd

import (
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
// if "bash" is not found in $PATH.
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
