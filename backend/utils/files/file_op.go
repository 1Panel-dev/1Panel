package files

import (
	"context"
	"github.com/mholt/archiver/v4"
	"github.com/spf13/afero"
	"io"
	"io/fs"
	"os"
	"path/filepath"
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

func (f FileOp) CreateFile(dst string) error {
	if _, err := f.Fs.Create(dst); err != nil {
		return err
	}
	return nil
}

func (f FileOp) LinkFile(source string, dst string, isSymlink bool) error {
	if isSymlink {
		osFs := afero.OsFs{}
		return osFs.SymlinkIfPossible(source, dst)
	} else {
		return os.Link(source, dst)
	}
}

func (f FileOp) DeleteDir(dst string) error {
	return f.Fs.RemoveAll(dst)
}

func (f FileOp) Stat(dst string) bool {
	info, _ := f.Fs.Stat(dst)
	if info != nil {
		return true
	}
	return false
}

func (f FileOp) DeleteFile(dst string) error {
	return f.Fs.Remove(dst)
}

func (f FileOp) WriteFile(dst string, in io.Reader, mode fs.FileMode) error {
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

func (f FileOp) Chmod(dst string, mode fs.FileMode) error {
	return f.Fs.Chmod(dst, mode)
}

type CompressType string

const (
	Zip   CompressType = "zip"
	Gz    CompressType = "gz"
	Bz2   CompressType = "bz2"
	Tar   CompressType = "tar"
	TarGz CompressType = "tarGz"
	Xz    CompressType = "xz"
)

func getFormat(cType CompressType) archiver.CompressedArchive {
	format := archiver.CompressedArchive{}
	switch cType {
	case Tar:
		format.Archival = archiver.Tar{}
	case TarGz, Gz:
		format.Compression = archiver.Gz{}
		format.Archival = archiver.Tar{}
	case Zip:
		format.Archival = archiver.Zip{}
	case Bz2:
		format.Compression = archiver.Bz2{}
		format.Archival = archiver.Tar{}
	case Xz:
		format.Compression = archiver.Xz{}
		format.Archival = archiver.Tar{}
	}
	return format
}

func (f FileOp) Compress(srcRiles []string, dst string, name string, cType CompressType) error {
	format := getFormat(cType)

	fileMaps := make(map[string]string, len(srcRiles))
	for _, s := range srcRiles {
		base := filepath.Base(s)
		fileMaps[s] = base
	}

	files, err := archiver.FilesFromDisk(nil, fileMaps)
	if err != nil {
		return err
	}
	dstFile := filepath.Join(dst, name)
	out, err := f.Fs.Create(dstFile)
	if err != nil {
		return err
	}

	err = format.Archive(context.Background(), out, files)
	if err != nil {
		return err
	}
	return nil
}

func (f FileOp) Decompress(srcFile string, dst string, cType CompressType) error {
	format := getFormat(cType)

	handler := func(ctx context.Context, archFile archiver.File) error {
		info := archFile.FileInfo
		filePath := filepath.Join(dst, archFile.NameInArchive)
		if archFile.FileInfo.IsDir() {
			if err := f.Fs.MkdirAll(filePath, info.Mode()); err != nil {
				return err
			}
			return nil
		}
		fr, err := archFile.Open()
		if err != nil {
			return err
		}
		defer fr.Close()
		fw, err := f.Fs.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, info.Mode())
		if err != nil {
			return err
		}
		defer fw.Close()
		if _, err := io.Copy(fw, fr); err != nil {
			return err
		}

		return nil
	}
	input, err := f.Fs.Open(srcFile)
	if err != nil {
		return err
	}
	return format.Extract(context.Background(), input, nil, handler)
}
