package files

import (
	"github.com/spf13/afero"
	"io"
	"io/fs"
	"os"
	"path"
)

type FileOp struct {
	Fs afero.Fs
}

func NewFileOp() FileOp {
	return FileOp{
		Fs: afero.NewOsFs(),
	}
}

func (f FileOp) CreateDir(dst string, mode fs.FileMode) error {
	return f.Fs.MkdirAll(dst, mode)
}

func (f FileOp) WriteFile(dst string, in io.Reader, mode fs.FileMode) error {
	dir, _ := path.Split(dst)
	if err := f.Fs.MkdirAll(dir, mode); err != nil {
		return err
	}
	file, err := f.Fs.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err = io.Copy(file, in); err != nil {
		return err
	}

	if _, err = file.Stat(); err != nil {
		return err
	}
	return nil
}
