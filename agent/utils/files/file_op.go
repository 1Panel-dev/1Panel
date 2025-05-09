package files

import (
	"archive/zip"
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/buserr"

	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/req_helper"
	cZip "github.com/klauspost/compress/zip"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"

	"github.com/1Panel-dev/1Panel/agent/global"
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

func (f FileOp) GetContent(dst string) ([]byte, error) {
	afs := &afero.Afero{Fs: f.Fs}
	cByte, err := afs.ReadFile(dst)
	if err != nil {
		return nil, err
	}
	return cByte, nil
}

func (f FileOp) CreateDir(dst string, mode fs.FileMode) error {
	return f.Fs.MkdirAll(dst, mode)
}

func (f FileOp) CreateDirWithMode(dst string, mode fs.FileMode) error {
	if err := f.Fs.MkdirAll(dst, mode); err != nil {
		return err
	}
	return f.ChmodRWithMode(dst, mode, true)
}
func (f FileOp) CreateDirWithPath(isDir bool, pathItem string) (string, error) {
	checkPath := pathItem
	if !isDir {
		checkPath = path.Dir(pathItem)
	}
	if !f.Stat(checkPath) {
		if err := f.CreateDir(checkPath, os.ModePerm); err != nil {
			return pathItem, err
		}
	}
	return pathItem, nil
}

func (f FileOp) CreateFile(dst string) error {
	if _, err := f.Fs.Create(dst); err != nil {
		return err
	}
	return nil
}

func (f FileOp) CreateFileWithMode(dst string, mode fs.FileMode) error {
	file, err := f.Fs.OpenFile(dst, os.O_CREATE, mode)
	if err != nil {
		return err
	}
	return file.Close()
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

func (f FileOp) CleanDir(dst string) error {
	return cmd.RunDefaultBashCf("rm -rf %s/*", dst)
}

func (f FileOp) RmRf(dst string) error {
	return cmd.RunDefaultBashCf("rm -rf %s", dst)
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

func (f FileOp) SaveFile(dst string, content string, mode fs.FileMode) error {
	if !f.Stat(path.Dir(dst)) {
		_ = f.CreateDir(path.Dir(dst), mode.Perm())
	}
	file, err := f.Fs.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	_, _ = write.WriteString(content)
	write.Flush()
	return nil
}

func (f FileOp) SaveFileWithByte(dst string, content []byte, mode fs.FileMode) error {
	if !f.Stat(path.Dir(dst)) {
		_ = f.CreateDir(path.Dir(dst), mode.Perm())
	}
	file, err := f.Fs.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	_, _ = write.Write(content)
	write.Flush()
	return nil
}

func (f FileOp) ChownR(dst string, uid string, gid string, sub bool) error {
	cmdStr := fmt.Sprintf(`%s chown %s:%s "%s"`, cmd.SudoHandleCmd(), uid, gid, dst)
	if sub {
		cmdStr = fmt.Sprintf(`chown -R %s:%s "%s"`, uid, gid, dst)
	}
	cmdMgr := cmd.NewCommandMgr(cmd.WithTimeout(10 * time.Second))
	if msg, err := cmdMgr.RunWithStdoutBashC(cmdStr); err != nil {
		if msg != "" {
			return errors.New(msg)
		}
		return err
	}
	return nil
}

func (f FileOp) ChmodR(dst string, mode int64, sub bool) error {
	cmdStr := fmt.Sprintf(`%s chmod %v "%s"`, cmd.SudoHandleCmd(), fmt.Sprintf("%04o", mode), dst)
	if sub {
		cmdStr = fmt.Sprintf(`%s chmod -R %v "%s"`, cmd.SudoHandleCmd(), fmt.Sprintf("%04o", mode), dst)
	}
	cmdMgr := cmd.NewCommandMgr(cmd.WithTimeout(10 * time.Second))
	if msg, err := cmdMgr.RunWithStdoutBashC(cmdStr); err != nil {
		if msg != "" {
			return errors.New(msg)
		}
		return err
	}
	return nil
}

func (f FileOp) ChmodRWithMode(dst string, mode fs.FileMode, sub bool) error {
	cmdStr := fmt.Sprintf(`%s chmod %v "%s"`, cmd.SudoHandleCmd(), fmt.Sprintf("%o", mode.Perm()), dst)
	if sub {
		cmdStr = fmt.Sprintf(`%s chmod -R %v "%s"`, cmd.SudoHandleCmd(), fmt.Sprintf("%o", mode.Perm()), dst)
	}
	cmdMgr := cmd.NewCommandMgr(cmd.WithTimeout(10 * time.Second))
	if msg, err := cmdMgr.RunWithStdoutBashC(cmdStr); err != nil {
		if msg != "" {
			return errors.New(msg)
		}
		return err
	}
	return nil
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
	percentValue := 0.0
	if w.Total > 0 {
		percent := float64(w.Written) / float64(w.Total) * 100
		percentValue, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", percent), 64)
	}
	process := Process{
		Total:   w.Total,
		Written: w.Written,
		Percent: percentValue,
		Name:    w.Name,
	}
	by, _ := json.Marshal(process)
	if percentValue < 100 {
		global.CACHE.Set(w.Key, string(by))
	} else {
		global.CACHE.SetWithTTL(w.Key, string(by), time.Second*time.Duration(10))
	}
}

func (f FileOp) DownloadFileWithProcess(url, dst, key string, ignoreCertificate bool) error {
	client := &http.Client{}
	if ignoreCertificate {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
	defer client.CloseIdleConnections()
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil
	}
	request.Header.Set("Accept-Encoding", "identity")
	resp, err := client.Do(request)
	if err != nil {
		global.LOG.Errorf("get download file [%s] error, err %s", dst, err.Error())
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		global.LOG.Errorf("create download file [%s] error, err %s", dst, err.Error())
		return err
	}
	go func() {
		counter := &WriteCounter{}
		counter.Key = key
		if resp.ContentLength > 0 {
			counter.Total = uint64(resp.ContentLength)
		}
		counter.Name = filepath.Base(dst)
		if _, err = io.Copy(out, io.TeeReader(resp.Body, counter)); err != nil {
			global.LOG.Errorf("save download file [%s] error, err %s", dst, err.Error())
		}
		out.Close()
		resp.Body.Close()

		value := global.CACHE.Get(counter.Key)
		process := &Process{}
		_ = json.Unmarshal([]byte(value), process)
		process.Percent = 100
		process.Name = counter.Name
		process.Total = process.Written
		by, _ := json.Marshal(process)
		global.CACHE.Set(counter.Key, string(by))
	}()
	return nil
}

func (f FileOp) DownloadFile(url, dst string) error {
	resp, err := req_helper.HandleGet(url)
	if err != nil {
		return err
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
	return nil
}

func (f FileOp) Cut(oldPaths []string, dst, name string, cover bool) error {
	for _, p := range oldPaths {
		var dstPath string
		if name != "" {
			dstPath = filepath.Join(dst, name)
			if f.Stat(dstPath) {
				dstPath = dst
			}
		} else {
			base := filepath.Base(p)
			dstPath = filepath.Join(dst, base)
		}
		coverFlag := ""
		if cover {
			coverFlag = "-f"
		}

		if err := cmd.RunDefaultBashCf(`mv %s '%s' '%s'`, coverFlag, p, dstPath); err != nil {
			return err
		}
	}
	return nil
}

func (f FileOp) Mv(oldPath, dstPath string) error {
	if err := cmd.RunDefaultBashCf(`mv '%s' '%s'`, oldPath, dstPath); err != nil {
		return err
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

func (f FileOp) CopyAndReName(src, dst, name string, cover bool) error {
	if src = path.Clean("/" + src); src == "" {
		return os.ErrNotExist
	}
	if dst = path.Clean("/" + dst); dst == "" {
		return os.ErrNotExist
	}
	if src == "/" || dst == src {
		return os.ErrInvalid
	}

	srcInfo, err := f.Fs.Stat(src)
	if err != nil {
		return err
	}

	if name != "" && !cover {
		if f.Stat(filepath.Join(dst, name)) {
			return buserr.New("ErrFileIsExist")
		}
	}

	if srcInfo.IsDir() {
		dstPath := dst
		if name != "" && !cover {
			dstPath = filepath.Join(dst, name)
		}
		return cmd.RunDefaultBashCf(`cp -rf '%s' '%s'`, src, dstPath)
	} else {
		dstPath := filepath.Join(dst, name)
		if cover {
			dstPath = dst
		}
		return cmd.RunDefaultBashCf(`cp -f '%s' '%s'`, src, dstPath)
	}
}

func (f FileOp) CopyDirWithNewName(src, dst, newName string) error {
	dstDir := filepath.Join(dst, newName)
	return cmd.RunDefaultBashCf(`cp -rf '%s' '%s'`, src, dstDir)
}

func (f FileOp) CopyDir(src, dst string) error {
	srcInfo, err := f.Fs.Stat(src)
	if err != nil {
		return err
	}
	dstDir := filepath.Join(dst, srcInfo.Name())
	if err = f.Fs.MkdirAll(dstDir, srcInfo.Mode()); err != nil {
		return err
	}
	return cmd.RunDefaultBashCf(`cp -rf '%s' '%s'`, src, dst+"/")
}

func (f FileOp) CopyDirWithExclude(src, dst string, excludeNames []string) error {
	srcInfo, err := f.Fs.Stat(src)
	if err != nil {
		return err
	}
	dstDir := filepath.Join(dst, srcInfo.Name())
	if err = f.Fs.MkdirAll(dstDir, srcInfo.Mode()); err != nil {
		return err
	}
	if len(excludeNames) == 0 {
		return cmd.RunDefaultBashCf(`cp -rf '%s' '%s'`, src, dst+"/")
	}
	tmpFiles, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, item := range tmpFiles {
		isExclude := false
		for _, name := range excludeNames {
			if item.Name() == name {
				isExclude = true
				break
			}
		}
		if isExclude {
			continue
		}
		if item.IsDir() {
			if err := f.CopyDir(path.Join(src, item.Name()), dstDir); err != nil {
				return err
			}
			continue
		}
		if err := f.CopyFile(path.Join(src, item.Name()), dstDir); err != nil {
			return err
		}
	}

	return nil
}

func (f FileOp) CopyFile(src, dst string) error {
	dst = filepath.Clean(dst) + string(filepath.Separator)
	return cmd.RunDefaultBashCf(`cp -f '%s' '%s'`, src, dst+"/")
}

func (f FileOp) GetDirSize(path string) (int64, error) {
	duCmd := exec.Command("du", "-s", path)
	output, err := duCmd.Output()
	if err == nil {
		fields := strings.Fields(string(output))
		if len(fields) == 2 {
			var cmdSize int64
			_, err = fmt.Sscanf(fields[0], "%d", &cmdSize)
			if err == nil {
				return cmdSize * 1024, nil
			}
		}
	}

	var size int64
	err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return size, nil
}

func getFormat(cType CompressType) archiver.CompressedArchive {
	format := archiver.CompressedArchive{}
	switch cType {
	case Tar:
		format.Archival = archiver.Tar{}
	case TarGz, Gz:
		format.Compression = archiver.Gz{}
		format.Archival = archiver.Tar{}
	case SdkTarGz:
		format.Compression = archiver.Gz{}
		format.Archival = archiver.Tar{}
	case SdkZip, Zip:
		format.Archival = archiver.Zip{
			Compression: zip.Deflate,
		}
	case Bz2:
		format.Compression = archiver.Bz2{}
		format.Archival = archiver.Tar{}
	case Xz:
		format.Compression = archiver.Xz{}
		format.Archival = archiver.Tar{}
	}
	return format
}

func (f FileOp) Compress(srcRiles []string, dst string, name string, cType CompressType, secret string) error {
	format := getFormat(cType)

	fileMaps := make(map[string]string, len(srcRiles))
	for _, s := range srcRiles {
		base := filepath.Base(s)
		fileMaps[s] = base
	}

	if !f.Stat(dst) {
		_ = f.CreateDir(dst, constant.DirPerm)
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

	switch cType {
	case Zip:
		if err := ZipFile(files, out); err == nil {
			return nil
		}
		_ = f.DeleteFile(dstFile)
		return NewZipArchiver().Compress(srcRiles, dstFile, "")
	case TarGz:
		err = NewTarGzArchiver().Compress(srcRiles, dstFile, secret)
		if err != nil {
			_ = f.DeleteFile(dstFile)
			return err
		}
	default:
		err = format.Archive(context.Background(), out, files)
		if err != nil {
			_ = f.DeleteFile(dstFile)
			return err
		}
	}
	return nil
}

func isIgnoreFile(name string) bool {
	return strings.HasPrefix(name, "__MACOSX") || strings.HasSuffix(name, ".DS_Store") || strings.HasPrefix(name, "._")
}

func decodeGBK(input string) (string, error) {
	decoder := simplifiedchinese.GBK.NewDecoder()
	decoded, _, err := transform.String(decoder, input)
	if err != nil {
		return "", err
	}
	return decoded, nil
}

func (f FileOp) decompressWithSDK(srcFile string, dst string, cType CompressType) error {
	format := getFormat(cType)
	handler := func(ctx context.Context, archFile archiver.File) error {
		info := archFile.FileInfo
		if isIgnoreFile(archFile.Name()) {
			return nil
		}
		fileName := archFile.NameInArchive
		var err error
		if header, ok := archFile.Header.(cZip.FileHeader); ok {
			if header.NonUTF8 && header.Flags == 0 {
				fileName, err = decodeGBK(fileName)
				if err != nil {
					return err
				}
			}
		}
		filePath := filepath.Join(dst, fileName)
		if archFile.FileInfo.IsDir() {
			if err := f.Fs.MkdirAll(filePath, info.Mode()); err != nil {
				return err
			}
			return nil
		} else {
			parentDir := path.Dir(filePath)
			if !f.Stat(parentDir) {
				if err := f.Fs.MkdirAll(parentDir, info.Mode()); err != nil {
					return err
				}
			}
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

func (f FileOp) Decompress(srcFile string, dst string, cType CompressType, secret string) error {
	if cType == Tar || cType == Zip || cType == TarGz {
		shellArchiver, err := NewShellArchiver(cType)
		if !f.Stat(dst) {
			_ = f.CreateDir(dst, 0755)
		}
		if err == nil {
			if err = shellArchiver.Extract(srcFile, dst, secret); err == nil {
				return nil
			}
		}
	}
	return f.decompressWithSDK(srcFile, dst, cType)
}

func ZipFile(files []archiver.File, dst afero.File) error {
	zw := zip.NewWriter(dst)
	defer zw.Close()

	for _, file := range files {
		hdr, err := zip.FileInfoHeader(file)
		if err != nil {
			return err
		}
		hdr.Method = zip.Deflate
		hdr.Name = file.NameInArchive
		if file.IsDir() {
			if !strings.HasSuffix(hdr.Name, "/") {
				hdr.Name += "/"
			}
		}
		w, err := zw.CreateHeader(hdr)
		if err != nil {
			return err
		}
		if file.IsDir() {
			continue
		}

		if file.LinkTarget != "" {
			_, err = w.Write([]byte(filepath.ToSlash(file.LinkTarget)))
			if err != nil {
				return err
			}
		} else {
			fileReader, err := file.Open()
			if err != nil {
				return err
			}
			_, err = io.Copy(w, fileReader)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (f FileOp) TarGzCompressPro(withDir bool, src, dst, secret, exclusionRules string) error {
	if !f.Stat(path.Dir(dst)) {
		if err := f.Fs.MkdirAll(path.Dir(dst), constant.FilePerm); err != nil {
			return err
		}
	}
	workdir := src
	srcItem := "."
	if withDir {
		workdir = path.Dir(src)
		srcItem = path.Base(src)
	}
	commands := ""

	exMap := make(map[string]struct{})
	exStr := ""
	excludes := strings.Split(exclusionRules, ";")
	for _, exclude := range excludes {
		if len(exclude) == 0 {
			continue
		}
		if _, ok := exMap[exclude]; ok {
			continue
		}
		exStr += " --exclude "
		exStr += exclude
		exMap[exclude] = struct{}{}
	}

	itemPrefix := filepath.Base(src)
	if itemPrefix == "/" {
		itemPrefix = ""
	}
	if len(secret) != 0 {
		commands = fmt.Sprintf("tar --warning=no-file-changed --ignore-failed-read %s --exclude-from=<(find %s -type s -printf '%s' | sed 's|^|%s/|') -zcf - %s | openssl enc -aes-256-cbc -salt -k '%s' -out %s", exStr, src, "%P\n", itemPrefix, srcItem, secret, dst)
		global.LOG.Debug(strings.ReplaceAll(commands, fmt.Sprintf(" %s ", secret), "******"))
	} else {
		commands = fmt.Sprintf("tar --warning=no-file-changed --ignore-failed-read --exclude-from=<(find %s -type s -printf '%s' | sed 's|^|%s/|') -zcf %s %s %s", src, "%P\n", itemPrefix, dst, exStr, srcItem)
		global.LOG.Debug(commands)
	}

	cmdMgr := cmd.NewCommandMgr(cmd.WithWorkDir(workdir))
	return cmdMgr.RunBashC(commands)
}

func (f FileOp) TarGzFilesWithCompressPro(list []string, dst, secret string) error {
	if !f.Stat(path.Dir(dst)) {
		if err := f.Fs.MkdirAll(path.Dir(dst), constant.FilePerm); err != nil {
			return err
		}
	}

	var filelist []string
	for _, item := range list {
		filelist = append(filelist, "-C '"+path.Dir(item)+"' '"+path.Base(item)+"' ")
	}
	commands := ""
	if len(secret) != 0 {
		commands = fmt.Sprintf("tar --warning=no-file-changed --ignore-failed-read -zcf - %s | openssl enc -aes-256-cbc -salt -k '%s' -out %s", strings.Join(filelist, " "), secret, dst)
		global.LOG.Debug(strings.ReplaceAll(commands, fmt.Sprintf(" %s ", secret), "******"))
	} else {
		commands = fmt.Sprintf("tar --warning=no-file-changed --ignore-failed-read -zcf %s %s", dst, strings.Join(filelist, " "))
		global.LOG.Debug(commands)
	}
	return cmd.RunDefaultBashC(commands)
}

func (f FileOp) TarGzExtractPro(src, dst string, secret string) error {
	if _, err := os.Stat(path.Dir(dst)); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(path.Dir(dst), os.ModePerm); err != nil {
			return err
		}
	}

	commands := ""
	if len(secret) != 0 {
		commands = fmt.Sprintf("openssl enc -d -aes-256-cbc -salt -k '%s' -in %s | tar -zxf - > /root/log", secret, src)
		global.LOG.Debug(strings.ReplaceAll(commands, fmt.Sprintf(" %s ", secret), "******"))
	} else {
		commands = fmt.Sprintf("tar zxvf %s", src)
		global.LOG.Debug(commands)
	}
	cmdMgr := cmd.NewCommandMgr(cmd.WithWorkDir(dst))
	return cmdMgr.RunBashC(commands)
}
func CopyCustomAppFile(srcPath, dstPath string) error {
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return fmt.Errorf("source file does not exist: %s", srcPath)
	}

	destDir := path.Dir(dstPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory %s: %v", destDir, err)
	}

	source, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %v", srcPath, err)
	}
	defer source.Close()

	tempFile, err := os.CreateTemp(destDir, "temp_*.tar.gz")
	if err != nil {
		return fmt.Errorf("failed to create temporary file in %s: %v", destDir, err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	if _, err = io.Copy(tempFile, source); err != nil {
		return fmt.Errorf("failed to copy file contents: %v", err)
	}

	tempFile.Close()
	source.Close()

	if err = os.Rename(tempFile.Name(), dstPath); err != nil {
		return fmt.Errorf("failed to rename temporary file to %s: %v", dstPath, err)
	}
	return nil
}
