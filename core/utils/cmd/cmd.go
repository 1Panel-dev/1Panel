package cmd

import (
	"os/exec"
)

func SudoHandleCmd() string {
	cmd := exec.Command("sudo", "-n", "ls")
	if err := cmd.Run(); err == nil {
		return "sudo "
	}
	return ""
}

func Which(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
