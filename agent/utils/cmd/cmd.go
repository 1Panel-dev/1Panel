package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
)

func Exec(cmdStr string) (string, error) {
	return ExecWithTimeOut(cmdStr, 20*time.Second)
}

func handleErr(stdout, stderr bytes.Buffer, err error) (string, error) {
	errMsg := ""
	if len(stderr.String()) != 0 {
		errMsg = fmt.Sprintf("stderr: %s", stderr.String())
	}
	if len(stdout.String()) != 0 {
		if len(errMsg) != 0 {
			errMsg = fmt.Sprintf("%s; stdout: %s", errMsg, stdout.String())
		} else {
			errMsg = fmt.Sprintf("stdout: %s", stdout.String())
		}
	}
	return errMsg, err
}

func ExecWithTimeOut(cmdStr string, timeout time.Duration) (string, error) {
	env := os.Environ()
	cmd := exec.Command("bash", "-c", cmdStr)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Env = env
	if err := cmd.Start(); err != nil {
		return "", err
	}
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	after := time.After(timeout)
	select {
	case <-after:
		_ = cmd.Process.Kill()
		return "", buserr.New("ErrCmdTimeout")
	case err := <-done:
		if err != nil {
			return handleErr(stdout, stderr, err)
		}
	}

	return stdout.String(), nil
}

func ExecWithLogFile(cmdStr string, timeout time.Duration, outputFile string) error {
	env := os.Environ()
	cmd := exec.Command("bash", "-c", cmdStr)

	outFile, err := os.OpenFile(outputFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, constant.DirPerm)
	if err != nil {
		return err
	}
	defer outFile.Close()

	cmd.Stdout = outFile
	cmd.Stderr = outFile
	cmd.Env = env

	if err := cmd.Start(); err != nil {
		return err
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	after := time.After(timeout)
	select {
	case <-after:
		_ = cmd.Process.Kill()
		return buserr.New("ErrCmdTimeout")
	case err := <-done:
		if err != nil {
			return err
		}
	}

	return nil
}

func ExecContainerScript(containerName, cmdStr string, timeout time.Duration) error {
	cmdStr = fmt.Sprintf("docker exec -i %s bash -c '%s'", containerName, cmdStr)
	out, err := ExecWithTimeOut(cmdStr, timeout)
	if err != nil {
		if out != "" {
			return fmt.Errorf("%s; err: %v", out, err)
		}
		return err
	}
	return nil
}

func ExecShell(outPath string, timeout time.Duration, name string, arg ...string) error {
	env := os.Environ()
	file, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE, constant.FilePerm)
	if err != nil {
		return err
	}
	defer file.Close()

	cmd := exec.Command(name, arg...)
	cmd.Stdout = file
	cmd.Stderr = file
	cmd.Env = env
	if err := cmd.Start(); err != nil {
		return err
	}
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	after := time.After(timeout)
	select {
	case <-after:
		_ = cmd.Process.Kill()
		return buserr.New("ErrCmdTimeout")
	case err := <-done:
		if err != nil {
			return err
		}
	}
	return nil
}

type CustomWriter struct {
	taskItem *task.Task
}

func (cw *CustomWriter) Write(p []byte) (n int, err error) {
	cw.taskItem.Log(string(p))
	return len(p), nil
}
func ExecShellWithTask(taskItem *task.Task, timeout time.Duration, name string, arg ...string) error {
	env := os.Environ()
	customWriter := &CustomWriter{taskItem: taskItem}
	cmd := exec.Command(name, arg...)
	cmd.Stdout = customWriter
	cmd.Stderr = customWriter
	cmd.Env = env
	if err := cmd.Start(); err != nil {
		return err
	}
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	after := time.After(timeout)
	select {
	case <-after:
		_ = cmd.Process.Kill()
		return buserr.New("ErrCmdTimeout")
	case err := <-done:
		if err != nil {
			return err
		}
	}
	return nil
}

func Execf(cmdStr string, a ...interface{}) (string, error) {
	env := os.Environ()
	cmd := exec.Command("bash", "-c", fmt.Sprintf(cmdStr, a...))
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Env = env
	err := cmd.Run()
	if err != nil {
		return handleErr(stdout, stderr, err)
	}
	return stdout.String(), nil
}

func ExecWithCheck(name string, a ...string) (string, error) {
	env := os.Environ()
	cmd := exec.Command(name, a...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Env = env
	err := cmd.Run()
	if err != nil {
		return handleErr(stdout, stderr, err)
	}
	return stdout.String(), nil
}

func ExecScript(scriptPath, workDir string) (string, error) {
	env := os.Environ()
	cmd := exec.Command("bash", scriptPath)
	var stdout, stderr bytes.Buffer
	cmd.Dir = workDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Env = env
	if err := cmd.Start(); err != nil {
		return "", err
	}
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	after := time.After(10 * time.Minute)
	select {
	case <-after:
		_ = cmd.Process.Kill()
		return "", buserr.New("ErrCmdTimeout")
	case err := <-done:
		if err != nil {
			return handleErr(stdout, stderr, err)
		}
	}

	return stdout.String(), nil
}

func ExecCmd(cmdStr string) error {
	cmd := exec.Command("bash", "-c", cmdStr)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error : %v, output: %s", err, output)
	}
	return nil
}

func ExecCmdWithDir(cmdStr, workDir string) error {
	cmd := exec.Command("bash", "-c", cmdStr)
	cmd.Dir = workDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error : %v, output: %s", err, output)
	}
	return nil
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

func HasNoPasswordSudo() bool {
	cmd2 := exec.Command("sudo", "-n", "ls")
	err2 := cmd2.Run()
	return err2 == nil
}

func SudoHandleCmd() string {
	cmd := exec.Command("sudo", "-n", "ls")
	if err := cmd.Run(); err == nil {
		return "sudo "
	}
	return ""
}

func Which(name string) bool {
	stdout, err := Execf("which %s", name)
	if err != nil || (len(strings.ReplaceAll(stdout, "\n", "")) == 0) {
		return false
	}
	return true
}

func ExecShellWithTimeOut(cmdStr, workdir string, logger *log.Logger, timeout time.Duration) error {
	env := os.Environ()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "-c", cmdStr)
	cmd.Dir = workdir
	cmd.Stdout = logger.Writer()
	cmd.Stderr = logger.Writer()
	cmd.Env = env
	if err := cmd.Start(); err != nil {
		return err
	}
	err := cmd.Wait()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return buserr.New("ErrCmdTimeout")
	}
	return err
}
