package terminal

import (
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/creack/pty"
	"github.com/pkg/errors"
)

const (
	DefaultCloseSignal  = syscall.SIGINT
	DefaultCloseTimeout = 10 * time.Second
)

type LocalCommand struct {
	closeSignal  syscall.Signal
	closeTimeout time.Duration

	cmd *exec.Cmd
	pty *os.File

	closeOnce sync.Once
	closeErr  error
}

func NewCommand(name string, arg ...string) (*LocalCommand, error) {
	cmd := exec.Command(name, arg...)
	if term := os.Getenv("TERM"); term != "" {
		cmd.Env = append(os.Environ(), "TERM="+term)
	} else {
		cmd.Env = append(os.Environ(), "TERM=xterm")
	}
	homeDir, _ := os.UserHomeDir()
	cmd.Dir = homeDir

	pty, err := pty.Start(cmd)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to start command")
	}

	lcmd := &LocalCommand{
		closeSignal:  DefaultCloseSignal,
		closeTimeout: DefaultCloseTimeout,

		cmd: cmd,
		pty: pty,
	}

	return lcmd, nil
}

func (lcmd *LocalCommand) Read(p []byte) (n int, err error) {
	return lcmd.pty.Read(p)
}

func (lcmd *LocalCommand) Write(p []byte) (n int, err error) {
	return lcmd.pty.Write(p)
}

func (lcmd *LocalCommand) Close() error {
	lcmd.closeOnce.Do(func() {
		if lcmd.pty != nil {
			_, _ = lcmd.pty.Write([]byte{3})
			time.Sleep(50 * time.Millisecond)
			_, _ = lcmd.pty.Write([]byte{4})
			time.Sleep(50 * time.Millisecond)
			_, _ = lcmd.pty.Write([]byte("exit\n"))
			time.Sleep(50 * time.Millisecond)
		}
		if lcmd.cmd != nil && lcmd.cmd.Process != nil {
			_ = lcmd.cmd.Process.Signal(syscall.SIGTERM)
			time.Sleep(50 * time.Millisecond)
			_ = lcmd.cmd.Process.Kill()
		}
		if lcmd.pty != nil {
			lcmd.closeErr = lcmd.pty.Close()
		}
	})
	return lcmd.closeErr
}

func (lcmd *LocalCommand) WaitResult() error {
	return lcmd.cmd.Wait()
}

func (lcmd *LocalCommand) Signal(signal os.Signal) error {
	if lcmd.cmd == nil || lcmd.cmd.Process == nil {
		return os.ErrProcessDone
	}
	return lcmd.cmd.Process.Signal(signal)
}

func (lcmd *LocalCommand) ResizeTerminal(width int, height int) error {
	window := struct {
		row uint16
		col uint16
		x   uint16
		y   uint16
	}{
		uint16(height),
		uint16(width),
		0,
		0,
	}
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		lcmd.pty.Fd(),
		syscall.TIOCSWINSZ,
		uintptr(unsafe.Pointer(&window)),
	)
	if errno != 0 {
		return errno
	} else {
		return nil
	}
}

func (lcmd *LocalCommand) Wait(quitChan chan bool) {
	if err := lcmd.WaitResult(); err != nil {
		global.LOG.Errorf("ssh session wait failed, err: %v", err)
		setQuit(quitChan)
	}
	setQuit(quitChan)
}
