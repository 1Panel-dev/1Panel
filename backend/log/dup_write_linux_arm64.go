package log

import (
	"os"
	"runtime"
	"syscall"
)

var stdErrFileHandler *os.File

func dupWrite(file *os.File) error {
	stdErrFileHandler = file
	if err := syscall.Dup3(int(file.Fd()), int(os.Stderr.Fd()), 0); err != nil {
		return err
	}
	runtime.SetFinalizer(stdErrFileHandler, func(fd *os.File) {
		fd.Close()
	})
	return nil
}
