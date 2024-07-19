package log

import (
	"golang.org/x/sys/unix"
	"os"
	"runtime"
)

var stdErrFileHandler *os.File

func dupWrite(file *os.File) error {
	stdErrFileHandler = file
	if err := unix.Dup2(int(file.Fd()), int(os.Stderr.Fd())); err != nil {
		return err
	}
	runtime.SetFinalizer(stdErrFileHandler, func(fd *os.File) {
		fd.Close()
	})
	return nil
}
