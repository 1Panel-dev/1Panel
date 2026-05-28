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

	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
)

const maxStreamOutputCapture = 64 * 1024

type CommandHelper struct {
	context      context.Context
	workDir      string
	outputFile   string
	scriptPath   string
	stdin        io.Reader
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

func RunDockerExecWithStdout(timeout time.Duration, containerName string, args ...string) (string, error) {
	commandArgs := append([]string{"exec", containerName}, args...)
	return NewCommandMgr(WithTimeout(timeout)).RunWithStdout("docker", commandArgs...)
}
func RunDockerExec(timeout time.Duration, containerName string, args ...string) error {
	commandArgs := append([]string{"exec", containerName}, args...)
	return NewCommandMgr(WithTimeout(timeout)).Run("docker", commandArgs...)
}

func (c *CommandHelper) Run(name string, arg ...string) error {
	_, err := c.run(name, arg...)
	return err
}

func (c *CommandHelper) RunWithOptionalSudo(name string, arg ...string) error {
	commandName, commandArgs := WrapWithOptionalSudo(name, arg...)
	return c.Run(commandName, commandArgs...)
}

func (c *CommandHelper) RunWithStdout(name string, arg ...string) (string, error) {
	return c.run(name, arg...)
}

func (c *CommandHelper) RunWithOptionalSudoAndStdout(name string, arg ...string) (string, error) {
	commandName, commandArgs := WrapWithOptionalSudo(name, arg...)
	return c.RunWithStdout(commandName, commandArgs...)
}

func (c *CommandHelper) RunPipe(commands ...PipeCommand) (string, error) {
	if len(commands) == 0 {
		return "", nil
	}

	ctx, cancel, cmds := c.preparePipeCommands(commands)
	if cancel != nil {
		defer cancel()
	}

	customWriter := &CustomWriter{taskItem: c.taskItem}
	var outputFile *os.File
	stdout, stderr := &lockedBuffer{}, &lockedBuffer{}
	limitOutputCapture := c.taskItem != nil || c.logger != nil || len(c.outputFile) != 0
	if limitOutputCapture {
		stdout.limit = maxStreamOutputCapture
		stderr.limit = maxStreamOutputCapture
	}
	var pipeStderr io.Writer = stderr
	var lastStdout io.Writer = stdout
	var lastStderr io.Writer = stderr
	var streamWriter io.Writer
	if c.taskItem != nil {
		streamWriter = customWriter
	} else if c.logger != nil {
		streamWriter = c.logger.Writer()
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
		if closer, ok := streamWriter.(io.Closer); ok {
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

	runErr := c.pipeResultErr(ctx, waitPipeCommands(ctx, cmds))
	if runErr != nil {
		return handleErrString(stdout.String(), stderr.String(), c.IgnoreExist1, runErr)
	}
	return stdout.String(), nil
}

func (c *CommandHelper) RunPipeToFile(outputFile string, commands ...PipeCommand) (string, error) {
	if len(commands) == 0 {
		return "", nil
	}

	ctx, cancel, cmds := c.preparePipeCommands(commands)
	if cancel != nil {
		defer cancel()
	}

	file, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, constant.FilePerm)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()

	stderr := &lockedBuffer{limit: maxStreamOutputCapture}
	if err := connectPipeCommands(cmds, file, stderr, stderr); err != nil {
		return "", err
	}
	if err := startPipeCommands(cmds); err != nil {
		return handleErrString("", stderr.String(), c.IgnoreExist1, err)
	}

	runErr := c.pipeResultErr(ctx, waitPipeCommands(ctx, cmds))
	if runErr != nil {
		return handleErrString("", stderr.String(), c.IgnoreExist1, runErr)
	}
	return "", nil
}

func (c *CommandHelper) preparePipeCommands(commands []PipeCommand) (context.Context, context.CancelFunc, []*exec.Cmd) {
	ctx, cancel := c.pipeContext()
	cmds := c.buildPipeCommands(ctx, commands)
	if commands[0].Stdin != nil {
		cmds[0].Stdin = commands[0].Stdin
	}
	return ctx, cancel, cmds
}

func (c *CommandHelper) pipeResultErr(ctx context.Context, runErr error) error {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return buserr.New("ErrCmdTimeout")
	}
	if errors.Is(ctx.Err(), context.Canceled) {
		return buserr.New("ErrShutDown")
	}
	return runErr
}

func (c *CommandHelper) pipeContext() (context.Context, context.CancelFunc) {
	ctx := c.context
	if ctx == nil {
		ctx = context.Background()
	}
	if c.timeout == 0 {
		return ctx, nil
	}
	return context.WithTimeout(ctx, c.timeout)
}

func (c *CommandHelper) buildPipeCommands(ctx context.Context, commands []PipeCommand) []*exec.Cmd {
	cmds := make([]*exec.Cmd, 0, len(commands))
	for _, item := range commands {
		cmdItem := exec.CommandContext(ctx, item.Name, filterEmptyArgs(item.Args)...)
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
	var newContext context.Context
	var cancel context.CancelFunc
	var outputFile *os.File
	arg = filterEmptyArgs(arg)

	if c.timeout != 0 {
		if c.context == nil {
			newContext, cancel = context.WithTimeout(context.Background(), c.timeout)
		} else {
			newContext, cancel = context.WithTimeout(c.context, c.timeout)
		}
		defer cancel()
	} else if c.context != nil {
		newContext = c.context
	}

	if len(c.scriptPath) != 0 {
		if newContext != nil {
			cmd = exec.CommandContext(newContext, "bash", c.scriptPath)
		} else {
			cmd = exec.Command("bash", c.scriptPath)
		}
	} else if newContext != nil {
		cmd = exec.CommandContext(newContext, name, arg...)
	} else {
		cmd = exec.Command(name, arg...)
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	customWriter := &CustomWriter{taskItem: c.taskItem}
	var stdout, stderr bytes.Buffer
	var loggerCloser io.Closer
	if c.taskItem != nil {
		cmd.Stdout = io.MultiWriter(&stdout, customWriter)
		cmd.Stderr = io.MultiWriter(&stderr, customWriter)
	} else if c.logger != nil {
		streamWriter := c.logger.Writer()
		if closer, ok := streamWriter.(io.Closer); ok {
			loggerCloser = closer
		}
		cmd.Stdout = io.MultiWriter(&stdout, streamWriter)
		cmd.Stderr = io.MultiWriter(&stderr, streamWriter)
	} else if len(c.outputFile) != 0 {
		file, err := os.OpenFile(c.outputFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, constant.FilePerm)
		if err != nil {
			return "", err
		}
		outputFile = file
		cmd.Stdout = io.MultiWriter(&stdout, outputFile)
		cmd.Stderr = io.MultiWriter(&stderr, outputFile)
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
	if c.stdin != nil {
		cmd.Stdin = c.stdin
	}
	defer func() {
		if loggerCloser != nil {
			_ = loggerCloser.Close()
		}
		if outputFile != nil {
			_ = outputFile.Close()
		}
	}()

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("cmd start failed: %w", err)
	}
	if c.taskItem != nil {
		defer customWriter.Flush()
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	select {
	case err := <-done:
		if err != nil {
			return handleErr(&stdout, &stderr, c.IgnoreExist1, err)
		}
		return stdout.String(), nil
	case <-contextDone(newContext):
		if cmd.Process != nil && cmd.Process.Pid > 0 {
			killErr := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
			_ = killErr
		}
		var err error
		switch newContext.Err() {
		case context.DeadlineExceeded:
			err = buserr.New("ErrCmdTimeout")
		case context.Canceled:
			err = buserr.New("ErrShutDown")
		default:
			err = newContext.Err()
		}
		<-done
		return "", err
	}
}

func filterEmptyArgs(args []string) []string {
	if len(args) == 0 {
		return args
	}
	filtered := args[:0]
	for _, arg := range args {
		if arg == "" {
			continue
		}
		filtered = append(filtered, arg)
	}
	return filtered
}

func contextDone(ctx context.Context) <-chan struct{} {
	if ctx == nil {
		return nil
	}
	return ctx.Done()
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

func WithContext(ctx context.Context) Option {
	return func(s *CommandHelper) {
		s.context = ctx
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
func WithScriptPath(scriptPath string) Option {
	return func(s *CommandHelper) {
		s.scriptPath = scriptPath
	}
}
func WithStdin(stdin io.Reader) Option {
	return func(s *CommandHelper) {
		s.stdin = stdin
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
	mu       sync.Mutex
	taskItem *task.Task
	buffer   bytes.Buffer
}

func (cw *CustomWriter) Write(p []byte) (n int, err error) {
	cw.mu.Lock()
	defer cw.mu.Unlock()
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
	cw.mu.Lock()
	defer cw.mu.Unlock()
	if cw.buffer.Len() > 0 {
		cw.taskItem.Log(cw.buffer.String())
		cw.buffer.Reset()
	}
}

func handleErr(stdout, stderr fmt.Stringer, ignoreExist1 bool, err error) (string, error) {
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
