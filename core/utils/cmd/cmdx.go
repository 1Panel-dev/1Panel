package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/1Panel-dev/1Panel/core/app/task"
	"github.com/1Panel-dev/1Panel/core/buserr"
	"github.com/1Panel-dev/1Panel/core/constant"
)

const maxStreamOutputCapture = 64 * 1024

type CommandHelper struct {
	workDir      string
	outputFile   string
	env          []string
	timeout      time.Duration
	taskItem     *task.Task
	logger       *log.Logger
	IgnoreExist1 bool
}

type Option func(*CommandHelper)

type PipeCommand struct {
	Name  string
	Args  []string
	Env   []string
	Dir   string
	Stdin io.Reader
}

type lockedBuffer struct {
	mu        sync.Mutex
	buf       bytes.Buffer
	limit     int
	truncated int
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.limit > 0 && b.buf.Len() >= b.limit {
		b.truncated += len(p)
		return len(p), nil
	}
	if b.limit > 0 && b.buf.Len()+len(p) > b.limit {
		keep := b.limit - b.buf.Len()
		_, _ = b.buf.Write(p[:keep])
		b.truncated += len(p) - keep
		return len(p), nil
	}
	return b.buf.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.truncated == 0 {
		return b.buf.String()
	}
	return fmt.Sprintf("%s\n... truncated %d bytes ...", b.buf.String(), b.truncated)
}

func NewCommandMgr(opts ...Option) *CommandHelper {
	s := &CommandHelper{}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (c *CommandHelper) Run(name string, arg ...string) error {
	_, err := c.run(name, arg...)
	return err
}

func (c *CommandHelper) RunWithStdout(name string, arg ...string) (string, error) {
	return c.run(name, arg...)
}

func (c *CommandHelper) RunPipe(commands ...PipeCommand) (string, error) {
	if len(commands) == 0 {
		return "", nil
	}

	ctx, cancel := c.pipeContext()
	if cancel != nil {
		defer cancel()
	}

	cmds := c.buildPipeCommands(ctx, commands)
	customWriter := &CustomWriter{taskItem: c.taskItem}
	var outputFile *os.File
	limitOutputCapture := c.taskItem != nil || c.logger != nil || len(c.outputFile) != 0
	stdout, stderr := &lockedBuffer{}, &lockedBuffer{}
	if limitOutputCapture {
		stdout.limit = maxStreamOutputCapture
		stderr.limit = maxStreamOutputCapture
	}
	if commands[0].Stdin != nil {
		cmds[0].Stdin = commands[0].Stdin
	}
	var pipeStderr io.Writer = stderr
	var lastStdout io.Writer = stdout
	var lastStderr io.Writer = stderr
	var streamWriter io.Writer
	var streamClosers []io.Closer
	if c.taskItem != nil {
		streamWriter = customWriter
	} else if c.logger != nil {
		streamWriter = c.logger.Writer()
		if closer, ok := streamWriter.(io.Closer); ok {
			streamClosers = append(streamClosers, closer)
		}
	} else if len(c.outputFile) != 0 {
		file, err := os.OpenFile(c.outputFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, constant.FilePerm)
		if err != nil {
			return "", err
		}
		outputFile = file
		lastStdout = outputFile
	}
	if streamWriter != nil {
		pipeStderr = io.MultiWriter(stderr, streamWriter)
		lastStdout = io.MultiWriter(stdout, streamWriter)
		lastStderr = io.MultiWriter(stderr, streamWriter)
	}
	defer func() {
		if c.taskItem != nil {
			customWriter.Flush()
		}
		for _, closer := range streamClosers {
			_ = closer.Close()
		}
		if outputFile != nil {
			_ = outputFile.Close()
		}
	}()
	if err := connectPipeCommands(cmds, lastStdout, lastStderr, pipeStderr); err != nil {
		return "", err
	}

	if err := startPipeCommands(cmds); err != nil {
		return handleErrString(stdout.String(), stderr.String(), c.IgnoreExist1, err)
	}

	runErr := waitPipeCommands(ctx, cmds)
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return "", buserr.New("ErrCmdTimeout")
	}
	if runErr != nil {
		return handleErrString(stdout.String(), stderr.String(), c.IgnoreExist1, runErr)
	}
	return stdout.String(), nil
}

func (c *CommandHelper) pipeContext() (context.Context, context.CancelFunc) {
	ctx := context.Background()
	if c.timeout == 0 {
		return ctx, nil
	}
	return context.WithTimeout(ctx, c.timeout)
}

func (c *CommandHelper) buildPipeCommands(ctx context.Context, commands []PipeCommand) []*exec.Cmd {
	cmds := make([]*exec.Cmd, 0, len(commands))
	for _, item := range commands {
		cmdItem := exec.CommandContext(ctx, item.Name, item.Args...)
		cmdItem.Env = append(os.Environ(), c.env...)
		cmdItem.Env = append(cmdItem.Env, item.Env...)
		cmdItem.Dir = c.workDir
		if item.Dir != "" {
			cmdItem.Dir = item.Dir
		}
		cmdItem.SysProcAttr = &syscall.SysProcAttr{
			Setpgid: true,
		}
		cmds = append(cmds, cmdItem)
	}
	return cmds
}

func connectPipeCommands(cmds []*exec.Cmd, stdout, stderr, pipeStderr io.Writer) error {
	for i := 0; i < len(cmds)-1; i++ {
		pipe, err := cmds[i].StdoutPipe()
		if err != nil {
			return err
		}
		cmds[i+1].Stdin = pipe
		cmds[i].Stderr = pipeStderr
	}
	last := cmds[len(cmds)-1]
	last.Stdout = stdout
	last.Stderr = stderr
	return nil
}

func startPipeCommands(cmds []*exec.Cmd) error {
	for i := len(cmds) - 1; i >= 0; i-- {
		if err := cmds[i].Start(); err != nil {
			killStarted(cmds[i+1:])
			return err
		}
	}
	return nil
}

func waitPipeCommands(ctx context.Context, cmds []*exec.Cmd) error {
	done := make(chan error, 1)
	go func() {
		var runErr error
		for _, item := range cmds {
			if err := item.Wait(); err != nil && runErr == nil {
				runErr = err
			}
		}
		done <- runErr
	}()
	select {
	case runErr := <-done:
		return runErr
	case <-ctx.Done():
		killProcessGroups(cmds)
		return <-done
	}
}

func (c *CommandHelper) run(name string, arg ...string) (string, error) {
	var cmd *exec.Cmd
	var ctx context.Context
	var cancel context.CancelFunc

	if c.timeout != 0 {
		ctx, cancel = context.WithTimeout(context.Background(), c.timeout)
		defer cancel()
		cmd = exec.CommandContext(ctx, name, arg...)
	} else {
		cmd = exec.Command(name, arg...)
	}

	customWriter := &CustomWriter{taskItem: c.taskItem}
	var stdout, stderr bytes.Buffer
	var loggerClosers []io.Closer
	if c.taskItem != nil {
		cmd.Stdout = customWriter
		cmd.Stderr = customWriter
	} else if c.logger != nil {
		stdoutWriter := c.logger.Writer()
		stderrWriter := c.logger.Writer()
		if closer, ok := stdoutWriter.(io.Closer); ok {
			loggerClosers = append(loggerClosers, closer)
		}
		if closer, ok := stderrWriter.(io.Closer); ok {
			loggerClosers = append(loggerClosers, closer)
		}
		cmd.Stdout = stdoutWriter
		cmd.Stderr = stderrWriter
	} else if len(c.outputFile) != 0 {
		file, err := os.OpenFile(c.outputFile, os.O_WRONLY|os.O_CREATE, constant.FilePerm)
		if err != nil {
			return "", err
		}
		defer file.Close()
		cmd.Stdout = file
		cmd.Stderr = file
	} else {
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
	}
	env := os.Environ()
	env = append(env, c.env...)
	cmd.Env = env
	if len(c.workDir) != 0 {
		cmd.Dir = c.workDir
	}
	defer func() {
		for _, closer := range loggerClosers {
			_ = closer.Close()
		}
	}()

	if c.timeout != 0 {
		err := cmd.Run()
		if c.taskItem != nil {
			customWriter.Flush()
		}
		if ctx != nil && errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return "", buserr.New("ErrCmdTimeout")
		}
		if err != nil {
			return handleErr(stdout, stderr, c.IgnoreExist1, err)
		}
		return stdout.String(), nil
	}

	err := cmd.Run()
	if c.taskItem != nil {
		customWriter.Flush()
	}
	if err != nil {
		return handleErr(stdout, stderr, c.IgnoreExist1, err)
	}
	return stdout.String(), nil
}

func killStarted(cmds []*exec.Cmd) {
	killProcessGroups(cmds)
	for _, item := range cmds {
		if item.Process != nil {
			_ = item.Wait()
		}
	}
}

func killProcessGroups(cmds []*exec.Cmd) {
	for _, item := range cmds {
		if item.Process != nil {
			_ = syscall.Kill(-item.Process.Pid, syscall.SIGKILL)
		}
	}
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
func WithEnv(env ...string) Option {
	return func(s *CommandHelper) {
		s.env = append(s.env, env...)
	}
}
func WithIgnoreExist1() Option {
	return func(s *CommandHelper) {
		s.IgnoreExist1 = true
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

func handleErr(stdout, stderr bytes.Buffer, ignoreExist1 bool, err error) (string, error) {
	return handleErrString(stdout.String(), stderr.String(), ignoreExist1, err)
}

func handleErrString(stdout, stderr string, ignoreExist1 bool, err error) (string, error) {
	var exitError *exec.ExitError
	if ignoreExist1 && errors.As(err, &exitError) {
		if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
			if status.ExitStatus() == 1 {
				return "", nil
			}
		}
	}
	outItem := stdout
	errItem := stderr
	if len(errItem) != 0 && len(outItem) != 0 {
		return outItem, fmt.Errorf("stdout: %s; stderr: %s, err: %v", outItem, errItem, err)
	}
	if len(errItem) != 0 {
		return outItem, fmt.Errorf("stderr: %s, err: %v", errItem, err)
	}
	if len(outItem) != 0 {
		return outItem, fmt.Errorf("stdout: %s, err: %v", outItem, err)
	}
	return "", err
}
