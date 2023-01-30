package files

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/mholt/archiver/v4"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

type FileOp struct {
	Fs afero.Fs
}

func NewFileOp() FileOp {
	return FileOp{
		Fs: afero.NewOsFs(),
	}
}

func (f FileOp) OpenFile(dst string) (fs.File, error) {
	return f.Fs.Open(dst)
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
	return info != nil
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

func (f FileOp) Rename(oldName string, newName string) error {
	return f.Fs.Rename(oldName, newName)
}

type WriteCounter struct {
	Total   uint64
	Written uint64
	Key     string
	Name    string
}

type Process struct {
	Total   uint64  `json:"total"`
	Written uint64  `json:"written"`
	Percent float64 `json:"percent"`
	Name    string  `json:"name"`
}

func (w *WriteCounter) Write(p []byte) (n int, err error) {
	n = len(p)
	w.Written += uint64(n)
	w.SaveProcess()
	return n, nil
}

func (w *WriteCounter) SaveProcess() {
	percent := float64(w.Written) / float64(w.Total) * 100
	percentValue, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", percent), 64)

	process := Process{
		Total:   w.Total,
		Written: w.Written,
		Percent: percentValue,
		Name:    w.Name,
	}

	by, _ := json.Marshal(process)

	if percentValue < 100 {
		if err := global.CACHE.Set(w.Key, string(by)); err != nil {
			global.LOG.Errorf("save cache error, err %s", err.Error())
		}
	} else {
		if err := global.CACHE.SetWithTTL(w.Key, string(by), time.Second*time.Duration(10)); err != nil {
			global.LOG.Errorf("save cache error, err %s", err.Error())
		}
	}

}

func (f FileOp) DownloadFileWithProcess(url, dst, key string) error {
	resp, err := http.Get(url)
	if err != nil {
		global.LOG.Errorf("get download file [%s] error, err %s", dst, err.Error())
	}
	header, err := http.Head(url)
	if err != nil {
		global.LOG.Errorf("get download file [%s] error, err %s", dst, err.Error())
	}

	out, err := os.Create(dst)
	if err != nil {
		global.LOG.Errorf("create download file [%s] error, err %s", dst, err.Error())
	}

	go func() {
		counter := &WriteCounter{}
		counter.Key = key
		counter.Total = uint64(header.ContentLength)
		counter.Name = filepath.Base(dst)
		if _, err = io.Copy(out, io.TeeReader(resp.Body, counter)); err != nil {
			global.LOG.Errorf("save download file [%s] error, err %s", dst, err.Error())
		}
		out.Close()
		resp.Body.Close()
	}()

	return nil
}

func (f FileOp) DownloadFile(url, dst string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("get download file [%s] error, err %s", dst, err.Error())
	}
	defer resp.Body.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create download file [%s] error, err %s", dst, err.Error())
	}
	defer out.Close()

	if _, err = io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("save download file [%s] error, err %s", dst, err.Error())
	}
	out.Close()
	resp.Body.Close()

	return nil
}

func (f FileOp) Cut(oldPaths []string, dst string) error {
	for _, p := range oldPaths {
		base := filepath.Base(p)
		dstPath := filepath.Join(dst, base)
		if err := f.Fs.Rename(p, dstPath); err != nil {
			return err
		}
	}
	return nil
}

func (f FileOp) Copy(src, dst string) error {
	if src = path.Clean("/" + src); src == "" {
		return os.ErrNotExist
	}

	if dst = path.Clean("/" + dst); dst == "" {
		return os.ErrNotExist
	}

	if src == "/" || dst == "/" {
		return os.ErrInvalid
	}

	if dst == src {
		return os.ErrInvalid
	}

	info, err := f.Fs.Stat(src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return f.CopyDir(src, dst)
	}

	return f.CopyFile(src, dst)
}

func (f FileOp) CopyDir(src, dst string) error {
	srcInfo, err := f.Fs.Stat(src)
	if err != nil {
		return err
	}
	dstDir := filepath.Join(dst, srcInfo.Name())
	if err := f.Fs.MkdirAll(dstDir, srcInfo.Mode()); err != nil {
		return err
	}

	dir, _ := f.Fs.Open(src)
	obs, err := dir.Readdir(-1)
	if err != nil {
		return err
	}
	var errs []error

	for _, obj := range obs {
		fSrc := filepath.Join(src, obj.Name())
		if obj.IsDir() {
			err = f.CopyDir(fSrc, dstDir)
			if err != nil {
				errs = append(errs, err)
			}
		} else {
			err = f.CopyFile(fSrc, dstDir)
			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	var errString string
	for _, err := range errs {
		errString += err.Error() + "\n"
	}

	if errString != "" {
		return errors.New(errString)
	}

	return nil
}

func (f FileOp) CopyFile(src, dst string) error {
	srcFile, err := f.Fs.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	err = f.Fs.MkdirAll(filepath.Dir(dst), 0666)
	if err != nil {
		return err
	}

	srcInfo, err := f.Fs.Stat(src)
	if err != nil {
		return err
	}

	dstFile, err := f.Fs.OpenFile(path.Join(dst, srcInfo.Name()), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0775)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err = io.Copy(dstFile, srcFile); err != nil {
		return err
	}
	info, err := f.Fs.Stat(src)
	if err != nil {
		return err
	}
	if err = f.Fs.Chmod(dstFile.Name(), info.Mode()); err != nil {
		return err
	}

	return nil
}

func (f FileOp) GetDirSize(path string) (float64, error) {
	var m sync.Map
	var wg sync.WaitGroup

	wg.Add(1)
	go ScanDir(f.Fs, path, &m, &wg)
	wg.Wait()

	var dirSize float64
	m.Range(func(k, v interface{}) bool {
		dirSize = dirSize + v.(float64)
		return true
	})

	return dirSize, nil
}

type CompressType string

const (
	Zip   CompressType = "zip"
	Gz    CompressType = "gz"
	Bz2   CompressType = "bz2"
	Tar   CompressType = "tar"
	TarGz CompressType = "tar.gz"
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

	if !f.Stat(dst) {
		_ = f.CreateDir(dst, 0755)
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

func (f FileOp) Backup(srcFile string) (string, error) {
	backupPath := srcFile + "_bak"
	info, _ := f.Fs.Stat(backupPath)
	if info != nil {
		if info.IsDir() {
			_ = f.DeleteDir(backupPath)
		} else {
			_ = f.DeleteFile(backupPath)
		}
	}
	if err := f.Rename(srcFile, backupPath); err != nil {
		return backupPath, err
	}

	return backupPath, nil
}

func (f FileOp) CopyAndBackup(src string) (string, error) {
	backupPath := src + "_bak"
	info, _ := f.Fs.Stat(backupPath)
	if info != nil {
		if info.IsDir() {
			_ = f.DeleteDir(backupPath)
		} else {
			_ = f.DeleteFile(backupPath)
		}
	}
	_ = f.CreateDir(backupPath, 0755)
	if err := f.Copy(src, backupPath); err != nil {
		return backupPath, err
	}
	return backupPath, nil
}
