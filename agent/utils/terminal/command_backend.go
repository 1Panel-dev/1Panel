package terminal

import (
	"io"
	"syscall"
	"time"
)

type commandBackend struct {
	command  *LocalCommand
	readDone chan struct{}
}

func newCommandBackend(command *LocalCommand, out io.Writer) *commandBackend {
	b := &commandBackend{command: command, readDone: make(chan struct{})}
	go func() {
		defer close(b.readDone)
		_, _ = io.Copy(out, command)
	}()
	return b
}

func (b *commandBackend) Write(p []byte) (int, error) {
	return b.command.Write(p)
}

func (b *commandBackend) Resize(cols, rows int) error {
	return b.command.ResizeTerminal(cols, rows)
}

func (b *commandBackend) Wait() error {
	err := b.command.WaitResult()
	select {
	case <-b.readDone:
	case <-time.After(time.Second):
	}
	return err
}

func (b *commandBackend) Keepalive() error {
	return b.command.Signal(syscall.Signal(0))
}

func (b *commandBackend) Close() error {
	return b.command.Close()
}
