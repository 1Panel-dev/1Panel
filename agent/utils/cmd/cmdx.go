package cmd

import (
	"bytes"
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

type CommandHelper struct {
	workDir    string
	outputFile string
	scriptPath string
	timeout    time.Duration
	taskItem   *task.Task
	logger     *log.Logger
}

type Option func(*CommandHelper)

func NewCommandMgr(opts ...Option) *CommandHelper {
	s := &CommandHelper{}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func RunDefaultBashC(command string) error {
	mgr := NewCommandMgr()
	return mgr.RunBashC(command)
}
func RunDefaultBashCf(command string, arg ...interface{}) error {
	mgr := NewCommandMgr()
	return mgr.RunBashCf(command, arg...)
}
func RunDefaultWithStdoutBashC(command string) (string, error) {
	mgr := NewCommandMgr(WithTimeout(20 * time.Second))
	return mgr.RunWithStdoutBashC(command)
}
func RunDefaultWithStdoutBashCf(command string, arg ...interface{}) (string, error) {
	mgr := NewCommandMgr(WithTimeout(20 * time.Second))
	return mgr.RunWithStdoutBashCf(command, arg...)
}

func (c *CommandHelper) Run(name string, arg ...string) error {
	_, err := c.run(name, arg...)
	return err
}
func (c *CommandHelper) RunBashCWithArgs(arg ...string) error {
	arg = append([]string{"-c"}, arg...)
	_, err := c.run("bash", arg...)
	return err
}
func (c *CommandHelper) RunBashC(command string) error {
	_, err := c.run("bash", "-c", command)
	return err
}
func (c *CommandHelper) RunBashCf(command string, arg ...interface{}) error {
	_, err := c.run("bash", "-c", fmt.Sprintf(command, arg...))
	return err
}

func (c *CommandHelper) RunWithStdout(name string, arg ...string) (string, error) {
	return c.run(name, arg...)
}
func (c *CommandHelper) RunWithStdoutBashC(command string) (string, error) {
	return c.run("bash", "-c", command)
}
func (c *CommandHelper) RunWithStdoutBashCf(command string, arg ...interface{}) (string, error) {
	return c.run("bash", "-c", fmt.Sprintf(command, arg...))
}

func (c *CommandHelper) run(name string, arg ...string) (string, error) {
	cmd := exec.Command(name, arg...)

	customWriter := &CustomWriter{taskItem: c.taskItem}
	var stdout, stderr bytes.Buffer
	if c.taskItem != nil {
		cmd.Stdout = customWriter
		cmd.Stderr = customWriter
	} else if c.logger != nil {
		cmd.Stdout = c.logger.Writer()
		cmd.Stderr = c.logger.Writer()
	} else if len(c.outputFile) != 0 {
		file, err := os.OpenFile(c.outputFile, os.O_WRONLY|os.O_CREATE, constant.FilePerm)
		if err != nil {
			return "", err
		}
		defer file.Close()
		cmd.Stdout = file
		cmd.Stderr = file
	} else if len(c.scriptPath) != 0 {
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		cmd = exec.Command("bash", c.scriptPath)
	} else {
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
	}
	env := os.Environ()
	cmd.Env = env
	if len(c.workDir) != 0 {
		cmd.Dir = c.workDir
	}

	if err := cmd.Start(); err != nil {
		return "", err
	}
	if c.timeout != 0 {
		done := make(chan error, 1)
		go func() {
			done <- cmd.Wait()
			if c.taskItem != nil {
				customWriter.Flush()
			}
		}()
		after := time.After(c.timeout)
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

	err := cmd.Run()
	if err != nil {
		return handleErr(stdout, stderr, err)
	}
	return stdout.String(), nil
}

func WithOutputFile(outputFile string) Option {
	return func(s *CommandHelper) {
		s.outputFile = outputFile
	}
}
func WithTimeout(timeout time.Duration) Option {
	return func(s *CommandHelper) {
		s.timeout = timeout
	}
}
func WithLogger(logger *log.Logger) Option {
	return func(s *CommandHelper) {
		s.logger = logger
	}
}
func WithTask(taskItem task.Task) Option {
	return func(s *CommandHelper) {
		s.taskItem = &taskItem
	}
}
func WithWorkDir(workDir string) Option {
	return func(s *CommandHelper) {
		s.workDir = workDir
	}
}
func WithScriptPath(scriptPath string) Option {
	return func(s *CommandHelper) {
		s.scriptPath = scriptPath
	}
}

type CustomWriter struct {
	taskItem *task.Task
	buffer   bytes.Buffer
}

func (cw *CustomWriter) Write(p []byte) (n int, err error) {
	cw.buffer.Write(p)
	lines := strings.Split(cw.buffer.String(), "\n")

	for i := 0; i < len(lines)-1; i++ {
		cw.taskItem.Log(lines[i])
	}
	cw.buffer.Reset()
	cw.buffer.WriteString(lines[len(lines)-1])

	return len(p), nil
}
func (cw *CustomWriter) Flush() {
	if cw.buffer.Len() > 0 {
		cw.taskItem.Log(cw.buffer.String())
		cw.buffer.Reset()
	}
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
