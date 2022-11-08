package cmd

import (
	"bytes"
	"os/exec"
)

func Exec(cmdStr string) (string, error) {
	cmd := exec.Command("bash", "-c", cmdStr)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return string(stderr.Bytes()), err
	}
	return string(stdout.Bytes()), nil
}
