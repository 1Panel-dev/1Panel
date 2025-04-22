package terminal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func NewLocalTerminalInDir(cwd string, initCmd string) (*LocalCommand, error) {
	cmd := exec.Command("bash", "--noprofile", "--norc")

	if initCmd == "" {
		initCmd = `export PS1="\u@\h:\w\$ " && clear`
	}

	if cwd != "" {
		absPath, err := filepath.Abs(cwd)
		if err != nil {
			return nil, fmt.Errorf("invalid directory path: %s", cwd)
		}

		forbiddenDirs := []string{"/etc", "/root", "/boot", "/proc", "/sys"}
		for _, forbid := range forbiddenDirs {
			if strings.HasPrefix(absPath, forbid) {
				return nil, fmt.Errorf("access to directory '%s' is not allowed", absPath)
			}
		}

		cmd.Dir = absPath
	}

	term := os.Getenv("TERM")
	if term == "" {
		term = "xterm"
	}
	cmd.Env = append(os.Environ(), "TERM="+term)

	ptyFile, err := pty.Start(cmd)
	if err != nil {
		return nil, errors.Wrap(err, "failed to start terminal")
	}

	if initCmd != "" {
		time.Sleep(100 * time.Millisecond)
		_, _ = ptyFile.Write([]byte(initCmd + "\n"))
	}

	return &LocalCommand{
		closeSignal:  DefaultCloseSignal,
		closeTimeout: DefaultCloseTimeout,
		cmd:          cmd,
		pty:          ptyFile,
	}, nil
}

func (lcmd *LocalCommand) Read(p []byte) (n int, err error) {
	return lcmd.pty.Read(p)
}

func (lcmd *LocalCommand) Write(p []byte) (n int, err error) {
	return lcmd.pty.Write(p)
}

func (lcmd *LocalCommand) Close() error {
	if lcmd.cmd != nil && lcmd.cmd.Process != nil {
		_ = lcmd.cmd.Process.Kill()
	}
	_ = lcmd.pty.Close()
	return nil
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
	if err := lcmd.cmd.Wait(); err != nil {
		global.LOG.Errorf("ssh session wait failed, err: %v", err)
		setQuit(quitChan)
	}
	setQuit(quitChan)
}
