package cmd

import (
	"os/exec"
	"strings"
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
