package files

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
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
	"github.com/spf13/afero"
	"golang.org/x/sync/singleflight"
)

const (
	cmdDefaultTimeout           = 10 * time.Second
	cmdRecursiveTimeout         = 5 * time.Minute
	maxArchiveSymlinkTargetSize = 4 * 1024
)

var protectedPaths = []string{
	"/",
	"/bin",
	"/sbin",
	"/etc",
	"/boot",
	"/usr",
	"/lib",
	"/lib64",
	"/dev",
	"/proc",
	"/sys",
	"/root",
}

var (
	dirSizeGroup   singleflight.Group
	dirSizeLimiter = make(chan struct{}, 2)
)

func IsProtected(path string) bool {
	real, err := filepath.EvalSymlinks(path)
	if err == nil {
		path = real
	}

	abs, err := filepath.Abs(path)
	if err == nil {
		path = abs
	}

	for _, p := range protectedPaths {
		if path == p {
			return true
		}
	}
	return false
}

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
	file, err := f.Fs.Create(dst)
	if err != nil {
		return err
	}
	return file.Close()
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
	if IsProtected(dst) {
		return buserr.New("ErrPathNotDelete")
	}
	return f.Fs.RemoveAll(dst)
}

func (f FileOp) Stat(dst string) bool {
	info, _ := f.Fs.Stat(dst)
	return info != nil
}

func (f FileOp) DeleteFile(dst string) error {
	if IsProtected(dst) {
		return buserr.New("ErrPathNotDelete")
	}
	return f.Fs.Remove(dst)
}

func (f FileOp) CleanDir(dst string) error {
	if IsProtected(dst) {
		return buserr.New("ErrPathNotDelete")
	}
	items, err := afero.ReadDir(f.Fs, dst)
	if err != nil {
		return err
	}
	for _, item := range items {
		if err := f.Fs.RemoveAll(filepath.Join(dst, item.Name())); err != nil {
			return err
		}
	}
	return nil
}

func (f FileOp) RmRf(dst string) error {
	if IsProtected(dst) {
		return buserr.New("ErrPathNotDelete")
	}
	return f.Fs.RemoveAll(dst)
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
	args := []string{uid + ":" + gid, dst}
	if sub {
		args = append([]string{"-R", uid + ":" + gid}, dst)
	}
	timeout := cmdDefaultTimeout
	if sub {
		timeout = cmdRecursiveTimeout
	}
	cmdMgr := cmd.NewCommandMgr(cmd.WithTimeout(timeout))
	if err := cmdMgr.RunWithOptionalSudo("chown", args...); err != nil {
		return err
	}
	return nil
}

func (f FileOp) ChmodR(dst string, mode int64, sub bool) error {
	args := []string{fmt.Sprintf("%04o", mode), dst}
	if sub {
		args = append([]string{"-R", fmt.Sprintf("%04o", mode)}, dst)
	}
	timeout := cmdDefaultTimeout
	if sub {
		timeout = cmdRecursiveTimeout
	}
	cmdMgr := cmd.NewCommandMgr(cmd.WithTimeout(timeout))
	if err := cmdMgr.RunWithOptionalSudo("chmod", args...); err != nil {
		return err
	}
	return nil
}

func (f FileOp) ChmodRWithMode(dst string, mode fs.FileMode, sub bool) error {
	args := []string{fmt.Sprintf("%o", mode.Perm()), dst}
	if sub {
		args = append([]string{"-R", fmt.Sprintf("%o", mode.Perm())}, dst)
	}
	timeout := cmdDefaultTimeout
	if sub {
		timeout = cmdRecursiveTimeout
	}
	cmdMgr := cmd.NewCommandMgr(cmd.WithTimeout(timeout))
	if err := cmdMgr.RunWithOptionalSudo("chmod", args...); err != nil {
		return err
	}
	return nil
}

func (f FileOp) ChownRPaths(paths []string, uid string, gid string, sub bool) error {
	if len(paths) == 0 {
		return nil
	}
	if len(paths) == 1 {
		return f.ChownR(paths[0], uid, gid, sub)
	}
	args := []string{uid + ":" + gid}
	if sub {
		args = append([]string{"-R", uid + ":" + gid}, paths...)
	} else {
		args = append(args, paths...)
	}
	timeout := cmdDefaultTimeout
	if sub {
		timeout = cmdRecursiveTimeout
	}
	cmdMgr := cmd.NewCommandMgr(cmd.WithTimeout(timeout))
	if err := cmdMgr.RunWithOptionalSudo("chown", args...); err != nil {
		return err
	}
	return nil
}

func (f FileOp) ChmodRPaths(paths []string, mode int64, sub bool) error {
	if len(paths) == 0 {
		return nil
	}
	if len(paths) == 1 {
		return f.ChmodR(paths[0], mode, sub)
	}
	modeStr := fmt.Sprintf("%04o", mode)
	args := []string{modeStr}
	if sub {
		args = append([]string{"-R", modeStr}, paths...)
	} else {
		args = append(args, paths...)
	}
	timeout := cmdDefaultTimeout
	if sub {
		timeout = cmdRecursiveTimeout
	}
	cmdMgr := cmd.NewCommandMgr(cmd.WithTimeout(timeout))
	if err := cmdMgr.RunWithOptionalSudo("chmod", args...); err != nil {
		return err
	}
	return nil
}

func (f FileOp) Rename(oldName string, newName string) error {
	return f.Fs.Rename(oldName, newName)
}

type downloadTask struct {
	resp *http.Response
	file *os.File
	dst  string
}

var (
	downloadMu    sync.Mutex
	downloadTasks = make(map[string]*downloadTask)
)

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

type DownloadProxyConfig struct {
	Type     string
	URL      string
	Port     string
	User     string
	Password string
}

type DownloadOptions struct {
	IgnoreCertificate bool
	Proxy             *DownloadProxyConfig
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

func buildDownloadProxyURL(proxy DownloadProxyConfig) (*url.URL, error) {
	proxyType := strings.TrimSpace(proxy.Type)
	proxyHost := strings.TrimSpace(proxy.URL)
	if proxyType == "" || proxyHost == "" {
		return nil, buserr.New("ErrWgetProxyNotConfigured")
	}
	if !strings.Contains(proxyHost, "://") {
		proxyHost = fmt.Sprintf("%s://%s", proxyType, proxyHost)
	}
	parsedURL, err := url.Parse(proxyHost)
	if err != nil {
		return nil, buserr.WithDetail("ErrWgetProxyInvalid", err.Error(), err)
	}
	if parsedURL.Scheme == "" {
		parsedURL.Scheme = proxyType
	}
	if parsedURL.Host == "" && parsedURL.Path != "" {
		parsedURL.Host = parsedURL.Path
		parsedURL.Path = ""
	}
	if parsedURL.Host == "" {
		return nil, buserr.New("ErrWgetProxyNotConfigured")
	}
	if strings.TrimSpace(proxy.Port) != "" && parsedURL.Port() == "" {
		parsedURL.Host = net.JoinHostPort(parsedURL.Hostname(), strings.TrimSpace(proxy.Port))
	}
	if proxy.User != "" && proxy.Password != "" {
		parsedURL.User = url.UserPassword(proxy.User, proxy.Password)
	} else if proxy.User != "" {
		parsedURL.User = url.User(proxy.User)
	}
	return parsedURL, nil
}

func newDownloadHTTPClient(options DownloadOptions) (*http.Client, error) {
	if !options.IgnoreCertificate && options.Proxy == nil {
		return &http.Client{}, nil
	}
	transport := &http.Transport{}
	if options.IgnoreCertificate {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	if options.Proxy != nil {
		proxyURL, err := buildDownloadProxyURL(*options.Proxy)
		if err != nil {
			return nil, err
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}
	return &http.Client{Transport: transport}, nil
}

func (f FileOp) DownloadFileWithProcess(url, dst, key string, options DownloadOptions) error {
	client, err := newDownloadHTTPClient(options)
	if err != nil {
		return err
	}
	defer client.CloseIdleConnections()

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return buserr.WithDetail("ErrWgetRemoteFailed", err.Error(), err)
	}
	request.Header.Set("Accept-Encoding", "identity")

	resp, err := client.Do(request)
	if err != nil {
		global.LOG.Errorf("get download file [%s] error, err %s", dst, err.Error())
		return buserr.WithDetail("ErrWgetRemoteFailed", err.Error(), err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64*1024))
		_ = resp.Body.Close()
		global.LOG.Errorf("wget remote returned non-success status %s for url %s", resp.Status, url)
		return buserr.WithDetail("ErrWgetRemoteFailed", resp.StatusCode, nil)
	}

	ct := strings.ToLower(resp.Header.Get("Content-Type"))
	dstExt := strings.ToLower(filepath.Ext(dst))
	if (strings.Contains(ct, "text/html") || strings.Contains(ct, "text/xml")) &&
		dstExt != ".html" && dstExt != ".htm" && dstExt != ".xml" && dstExt != ".svg" {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64*1024))
		_ = resp.Body.Close()
		detail := fmt.Sprintf("Content-Type: %s", ct)
		global.LOG.Errorf("wget got html/xml response for non-html file %s, url %s, %s", dst, url, detail)
		return buserr.WithDetail("ErrWgetInvalidContentType", detail, nil)
	}

	out, err := os.Create(dst)
	if err != nil {
		global.LOG.Errorf("create download file [%s] error, err %s", dst, err.Error())
		resp.Body.Close()
		return err
	}

	downloadMu.Lock()
	downloadTasks[key] = &downloadTask{
		resp: resp,
		file: out,
		dst:  dst,
	}
	downloadMu.Unlock()

	go func() {
		defer func() {
			out.Close()
			resp.Body.Close()

			downloadMu.Lock()
			delete(downloadTasks, key)
			downloadMu.Unlock()
		}()

		counter := &WriteCounter{}
		counter.Key = key
		if resp.ContentLength > 0 {
			counter.Total = uint64(resp.ContentLength)
		}
		counter.Name = filepath.Base(dst)

		if _, err := io.Copy(out, io.TeeReader(resp.Body, counter)); err != nil {
			global.LOG.Errorf("save download file [%s] error, err %s", dst, err.Error())
			global.CACHE.Del(counter.Key)
			return
		}

		value := global.CACHE.Get(counter.Key)
		if value == "" {
			return
		}
		process := &Process{}
		if err := json.Unmarshal([]byte(value), process); err != nil {
			return
		}
		process.Percent = 100
		process.Name = counter.Name
		process.Total = process.Written
		by, _ := json.Marshal(process)
		global.CACHE.Set(counter.Key, string(by))
	}()
	return nil
}

func CancelDownload(key string) {
	downloadMu.Lock()
	task, ok := downloadTasks[key]
	if !ok {
		downloadMu.Unlock()
		return
	}
	dst := task.dst
	downloadMu.Unlock()

	_ = task.file.Close()
	_ = task.resp.Body.Close()

	if dst != "" {
		_ = os.Remove(dst)
	}
	global.CACHE.Del(key)
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
	if len(oldPaths) == 0 {
		return nil
	}
	var dstPath string
	coverFlag := ""
	if name != "" {
		dstPath = filepath.Join(dst, name)
		if f.Stat(dstPath) {
			dstPath = dst
		}
		if cover {
			coverFlag = "-f"
		}
	} else {
		dstPath = dst
		coverFlag = "-f"
	}
	args := []string{}
	if coverFlag != "" {
		args = append(args, coverFlag)
	}
	args = append(args, oldPaths...)
	args = append(args, dstPath)
	if err := cmd.NewCommandMgr(cmd.WithTimeout(cmdRecursiveTimeout)).Run("mv", args...); err != nil {
		return err
	}
	return nil
}

func (f FileOp) Mv(oldPath, dstPath string) error {
	if err := cmd.NewCommandMgr(cmd.WithTimeout(cmdRecursiveTimeout)).Run("mv", oldPath, dstPath); err != nil {
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
	ctx, cancel := context.WithTimeout(context.Background(), cmdRecursiveTimeout)
	defer cancel()
	return f.CopyAndReNameWithContext(ctx, src, dst, name, cover)
}

func (f FileOp) CopyAndReNameWithContext(ctx context.Context, src, dst, name string, cover bool) error {
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
		return cmd.NewCommandMgr(cmd.WithContext(ctx)).Run("cp", "-rfp", src, dstPath)
	} else {
		dstPath := filepath.Join(dst, name)
		if cover {
			dstPath = dst
		}
		return cmd.NewCommandMgr(cmd.WithContext(ctx)).Run("cp", "-fp", src, dstPath)
	}
}

func (f FileOp) CopyDirWithNewName(src, dst, newName string) error {
	if newName == "." || newName == "" {
		return cmd.NewCommandMgr(cmd.WithTimeout(cmdRecursiveTimeout)).Run("cp", "-rfp", filepath.Clean(src)+"/.", dst)
	}
	dstDir := filepath.Join(dst, newName)
	return cmd.NewCommandMgr(cmd.WithTimeout(cmdRecursiveTimeout)).Run("cp", "-rfp", src, dstDir)
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
	return cmd.NewCommandMgr(cmd.WithIgnoreExist1()).Run("cp", "-rfp", src, dst+"/")
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
		return cmd.NewCommandMgr(cmd.WithIgnoreExist1()).Run("cp", "-rfp", src, dst+"/")
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
	return cmd.NewCommandMgr(cmd.WithIgnoreExist1()).Run("cp", "-fp", src, dst+"/")
}

func (f FileOp) GetDirSize(path string) (int64, error) {
	cleanPath := filepath.Clean(path)
	result, err, _ := dirSizeGroup.Do("single:"+cleanPath, func() (interface{}, error) {
		dirSizeLimiter <- struct{}{}
		defer func() {
			<-dirSizeLimiter
		}()
		return f.getDirSize(cleanPath)
	})
	if err != nil {
		return 0, err
	}
	return result.(int64), nil
}

func (f FileOp) getDirSize(path string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cmdRecursiveTimeout)
	defer cancel()
	duCmd := exec.CommandContext(ctx, "du", "-s", path)
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
	if ctx.Err() != nil {
		return 0, ctx.Err()
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

type DirSize struct {
	Path string `json:"path"`
	Size int64  `json:"size"`
}

func (f FileOp) GetDepthDirSize(path string) ([]DirSize, error) {
	cleanPath := filepath.Clean(path)
	result, err, _ := dirSizeGroup.Do("depth:"+cleanPath, func() (interface{}, error) {
		dirSizeLimiter <- struct{}{}
		defer func() {
			<-dirSizeLimiter
		}()
		return f.getDepthDirSize(cleanPath)
	})
	if err != nil {
		return nil, err
	}
	return result.([]DirSize), nil
}

func (f FileOp) getDepthDirSize(path string) ([]DirSize, error) {
	var result []DirSize
	sizeMap := make(map[string]int64)
	ctx, cancel := context.WithTimeout(context.Background(), cmdRecursiveTimeout)
	defer cancel()
	duCmd := exec.CommandContext(ctx, "du", "-k", "--max-depth=1", "--exclude=proc", path)
	output, err := duCmd.Output()
	if err == nil {
		parseDUOutput(output, sizeMap)
	} else if ctx.Err() != nil {
		return nil, ctx.Err()
	} else {
		calculateDirSizeFallback(path, sizeMap)
	}

	for dir, size := range sizeMap {
		result = append(result, DirSize{
			Path: dir,
			Size: size,
		})
	}

	return result, nil
}

func parseDUOutput(output []byte, sizeMap map[string]int64) {
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		sizeText, dir, ok := strings.Cut(strings.TrimSpace(line), "\t")
		if !ok {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}
			sizeText = fields[0]
			dir = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), sizeText))
		}
		if sizeKB, err := strconv.ParseInt(strings.TrimSpace(sizeText), 10, 64); err == nil {
			sizeMap[strings.TrimSpace(dir)] = sizeKB * 1024
		}
	}
}

func calculateDirSizeFallback(path string, sizeMap map[string]int64) {
	_ = filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			rel, err := filepath.Rel(path, p)
			if err != nil {
				return nil
			}
			parts := strings.Split(rel, string(os.PathSeparator))
			var topLevel string
			if len(parts) == 0 || parts[0] == "." {
				topLevel = path
			} else {
				topLevel = filepath.Join(path, parts[0])
			}
			sizeMap[topLevel] += info.Size()
		}
		return nil
	})
}

func getFormat(cType CompressType) archiver.CompressedArchive {
	format := archiver.CompressedArchive{}
	switch cType {
	case Tar:
		format.Archival = archiver.Tar{}
	case TarGz, Gz, Tgz:
		format.Compression = archiver.Gz{}
		format.Archival = archiver.Tar{}
	case SdkTarGz:
		format.Compression = archiver.Gz{}
		format.Archival = archiver.Tar{}
	case SdkZip, Zip:
		format.Archival = archiver.Zip{
			Compression: zip.Deflate,
		}
	case Bz2, TarBz2:
		format.Compression = archiver.Bz2{}
		format.Archival = archiver.Tar{}
	case Xz, TarXz:
		format.Compression = archiver.Xz{}
		format.Archival = archiver.Tar{}
	}
	return format
}

func (f FileOp) Compress(ctx context.Context, srcRiles []string, dst string, name string, cType CompressType, secret string, progress func(current, total int, message string)) error {
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

	switch cType {
	case Zip, SdkZip:
		out, err := f.Fs.Create(dstFile)
		if err != nil {
			return err
		}
		defer out.Close()
		if err := ZipFile(ctx, files, out, progress); err == nil {
			return nil
		}
		_ = f.DeleteFile(dstFile)
		return NewZipArchiver().Compress(ctx, srcRiles, dstFile, "")
	case Tar, Gz, Bz2, TarBz2, Tgz, Xz, TarXz:
		err = NewTarArchiver(cType).Compress(ctx, srcRiles, dstFile, secret)
		if err != nil {
			_ = f.DeleteFile(dstFile)
			return err
		}
	case TarGz:
		err = NewTarGzArchiver().Compress(ctx, srcRiles, dstFile, secret)
		if err != nil {
			_ = f.DeleteFile(dstFile)
			return err
		}
	case Rar:
		if err := checkCmdAvailability("rar"); err != nil {
			return err
		}
		err = NewRarArchiver().Compress(ctx, srcRiles, dstFile, secret)
		if err != nil {
			_ = f.DeleteFile(dstFile)
			return err
		}
	case X7z:
		if err := checkCmdAvailability("7z"); err != nil {
			return err
		}
		err = NewX7zArchiver().Compress(ctx, srcRiles, dstFile, secret)
		if err != nil {
			_ = f.DeleteFile(dstFile)
			return err
		}
	default:
		tmpFile, err := os.CreateTemp(dst, fmt.Sprintf("temp_*%s", filepath.Ext(name)))
		if err != nil {
			return err
		}
		success := false
		defer func() {
			_ = tmpFile.Close()
			if !success {
				_ = os.Remove(tmpFile.Name())
				_ = f.DeleteFile(dstFile)
			}
		}()

		err = format.Archive(ctx, &contextAwareWriter{ctx: ctx, writer: tmpFile}, files)
		if err != nil {
			return err
		}
		if err = tmpFile.Close(); err != nil {
			return err
		}
		if err = os.Rename(tmpFile.Name(), dstFile); err != nil {
			return err
		}
		success = true
	}
	return nil
}

type contextAwareWriter struct {
	ctx    context.Context
	writer io.Writer
}

func (w *contextAwareWriter) Write(p []byte) (int, error) {
	if err := w.ctx.Err(); err != nil {
		return 0, err
	}
	return w.writer.Write(p)
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

type DecompressOptions struct {
	PreserveOwner     bool
	AllowCLIReextract bool
}

type archiveOwnership struct {
	uid int
	gid int
}

func getArchiveOwnership(file archiver.File) (archiveOwnership, bool) {
	header, ok := getArchiveTarHeader(file)
	if ok && header.Uid >= 0 && header.Gid >= 0 {
		return archiveOwnership{uid: header.Uid, gid: header.Gid}, true
	}
	return getArchiveZipOwnership(file)
}

const zipUnixOwnershipExtraID = 0x7875

func getArchiveZipOwnership(file archiver.File) (archiveOwnership, bool) {
	var extra []byte
	switch header := file.Header.(type) {
	case *zip.FileHeader:
		if header != nil {
			extra = header.Extra
		}
	case zip.FileHeader:
		extra = header.Extra
	case *cZip.FileHeader:
		if header != nil {
			extra = header.Extra
		}
	case cZip.FileHeader:
		extra = header.Extra
	default:
		return archiveOwnership{}, false
	}
	for len(extra) >= 4 {
		fieldID := binary.LittleEndian.Uint16(extra[:2])
		fieldSize := int(binary.LittleEndian.Uint16(extra[2:4]))
		extra = extra[4:]
		if fieldSize > len(extra) {
			return archiveOwnership{}, false
		}
		field := extra[:fieldSize]
		extra = extra[fieldSize:]
		if fieldID != zipUnixOwnershipExtraID || len(field) < 4 || field[0] != 1 {
			continue
		}
		uidSize := int(field[1])
		if uidSize == 0 || uidSize > 4 || len(field) < 2+uidSize+1 {
			return archiveOwnership{}, false
		}
		uid := decodeZipOwnershipID(field[2 : 2+uidSize])
		gidSizeOffset := 2 + uidSize
		gidSize := int(field[gidSizeOffset])
		if gidSize == 0 || gidSize > 4 || len(field) < gidSizeOffset+1+gidSize {
			return archiveOwnership{}, false
		}
		gid := decodeZipOwnershipID(field[gidSizeOffset+1 : gidSizeOffset+1+gidSize])
		return archiveOwnership{uid: int(uid), gid: int(gid)}, true
	}
	return archiveOwnership{}, false
}

func decodeZipOwnershipID(value []byte) uint32 {
	var result uint32
	for i := len(value) - 1; i >= 0; i-- {
		result = result<<8 | uint32(value[i])
	}
	return result
}

func appendZipOwnershipExtra(extra []byte, ownership archiveOwnership) []byte {
	const fieldSize = 11
	field := make([]byte, 4+fieldSize)
	binary.LittleEndian.PutUint16(field[:2], zipUnixOwnershipExtraID)
	binary.LittleEndian.PutUint16(field[2:4], fieldSize)
	field[4] = 1
	field[5] = 4
	binary.LittleEndian.PutUint32(field[6:10], uint32(ownership.uid))
	field[10] = 4
	binary.LittleEndian.PutUint32(field[11:15], uint32(ownership.gid))
	return append(extra, field...)
}

func getFileOwnership(info fs.FileInfo) (archiveOwnership, bool) {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return archiveOwnership{}, false
	}
	return archiveOwnership{uid: int(stat.Uid), gid: int(stat.Gid)}, true
}

func getArchiveTarHeader(file archiver.File) (tar.Header, bool) {
	switch header := file.Header.(type) {
	case *tar.Header:
		if header == nil {
			return tar.Header{}, false
		}
		return *header, true
	case tar.Header:
		return header, true
	default:
		return tar.Header{}, false
	}
}

func applyArchiveOwnership(filePath string, mode fs.FileMode, ownership archiveOwnership) error {
	if mode&fs.ModeSymlink != 0 {
		return os.Lchown(filePath, ownership.uid, ownership.gid)
	}
	return os.Chown(filePath, ownership.uid, ownership.gid)
}

func archiveDestinationPath(dst, name string) (string, error) {
	cleaned := filepath.Clean(filepath.FromSlash(name))
	if cleaned == "." {
		return dst, nil
	}
	if filepath.IsAbs(cleaned) || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("invalid archive path: %s", name)
	}
	return filepath.Join(dst, cleaned), nil
}

func ensureArchiveDirectory(dirPath string, mode fs.FileMode) error {
	info, err := os.Lstat(dirPath)
	if err == nil {
		if info.Mode()&fs.ModeSymlink != 0 || !info.IsDir() {
			return fmt.Errorf("archive directory path is not a directory: %s", dirPath)
		}
		return nil
	}
	if !os.IsNotExist(err) {
		return err
	}
	return os.Mkdir(dirPath, mode.Perm())
}

func ensureArchiveParent(dst, filePath string) error {
	root := filepath.Clean(dst)
	parent := filepath.Dir(filePath)
	rel, err := filepath.Rel(root, parent)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("archive parent escapes destination: %s", parent)
	}
	if err := ensureArchiveDirectory(root, constant.DirPerm); err != nil {
		return err
	}
	if rel == "." {
		return nil
	}
	current := root
	for _, part := range strings.Split(rel, string(os.PathSeparator)) {
		current = filepath.Join(current, part)
		if err := ensureArchiveDirectory(current, constant.DirPerm); err != nil {
			return err
		}
	}
	return nil
}

func validateArchiveHardlinkTarget(dst, targetPath string) error {
	root := filepath.Clean(dst)
	rel, err := filepath.Rel(root, targetPath)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("archive hardlink target escapes destination: %s", targetPath)
	}
	current := root
	parts := strings.Split(rel, string(os.PathSeparator))
	for i, part := range parts {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if err != nil {
			return fmt.Errorf("archive hardlink target is unavailable: %s: %w", targetPath, err)
		}
		if info.Mode()&fs.ModeSymlink != 0 {
			return fmt.Errorf("archive hardlink target contains a symlink: %s", current)
		}
		if i < len(parts)-1 && !info.IsDir() {
			return fmt.Errorf("archive hardlink target parent is not a directory: %s", current)
		}
		if i == len(parts)-1 && !info.Mode().IsRegular() {
			return fmt.Errorf("archive hardlink target is not a regular file: %s", current)
		}
	}
	return nil
}

func (f FileOp) extractArchiveWithSDK(ctx context.Context, input io.Reader, dst string, extractor archiver.Extractor, options DecompressOptions) (bool, error) {
	type dirEntry struct {
		path         string
		mode         fs.FileMode
		modTime      time.Time
		ownership    archiveOwnership
		hasOwnership bool
	}
	type hardlinkEntry struct {
		path         string
		target       string
		mode         fs.FileMode
		modTime      time.Time
		ownership    archiveOwnership
		hasOwnership bool
	}
	var dirs []dirEntry
	var hardlinks []hardlinkEntry
	extractionStarted := false

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
		header, hasTarHeader := getArchiveTarHeader(archFile)
		ownership, hasOwnership := getArchiveOwnership(archFile)
		filePath, err := archiveDestinationPath(dst, fileName)
		if err != nil {
			return err
		}
		extractionStarted = true
		if archFile.FileInfo.IsDir() {
			if filePath != filepath.Clean(dst) {
				if err := ensureArchiveParent(dst, filePath); err != nil {
					return err
				}
			}
			if err := ensureArchiveDirectory(filePath, info.Mode()); err != nil {
				return err
			}
			dirs = append(dirs, dirEntry{
				path:         filePath,
				mode:         info.Mode(),
				modTime:      info.ModTime(),
				ownership:    ownership,
				hasOwnership: hasOwnership,
			})
			return nil
		}

		if hasTarHeader && header.Typeflag == tar.TypeLink {
			target, err := archiveDestinationPath(dst, header.Linkname)
			if err != nil {
				return err
			}
			hardlinks = append(hardlinks, hardlinkEntry{
				path:         filePath,
				target:       target,
				mode:         info.Mode(),
				modTime:      info.ModTime(),
				ownership:    ownership,
				hasOwnership: hasOwnership,
			})
			return nil
		}

		if err := ensureArchiveParent(dst, filePath); err != nil {
			return err
		}

		if info.Mode()&fs.ModeSymlink != 0 || hasTarHeader && header.Typeflag == tar.TypeSymlink {
			target := archFile.LinkTarget
			if target == "" && hasTarHeader {
				target = header.Linkname
			}
			if target == "" && archFile.Open != nil {
				fr, err := archFile.Open()
				if err != nil {
					return err
				}
				data, readErr := io.ReadAll(io.LimitReader(fr, maxArchiveSymlinkTargetSize+1))
				closeErr := fr.Close()
				if readErr != nil {
					return readErr
				}
				if closeErr != nil {
					return closeErr
				}
				target = string(data)
			}
			if len(target) > maxArchiveSymlinkTargetSize {
				return fmt.Errorf("archive symlink target for %s exceeds %d bytes", fileName, maxArchiveSymlinkTargetSize)
			}
			if target == "" {
				return fmt.Errorf("archive symlink %s has no target", fileName)
			}
			if err := os.RemoveAll(filePath); err != nil {
				return err
			}
			if err := os.Symlink(target, filePath); err != nil {
				return err
			}
			if options.PreserveOwner && hasOwnership {
				if err := applyArchiveOwnership(filePath, info.Mode(), ownership); err != nil {
					return fmt.Errorf("restore archive ownership for %s: %w", fileName, err)
				}
			}
			return nil
		}
		if existing, err := os.Lstat(filePath); err == nil {
			if existing.IsDir() {
				return fmt.Errorf("archive file path is a directory: %s", fileName)
			}
			if err := os.Remove(filePath); err != nil {
				return err
			}
		} else if !os.IsNotExist(err) {
			return err
		}

		fr, err := archFile.Open()
		if err != nil {
			return err
		}
		fw, err := f.Fs.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, info.Mode().Perm())
		if err != nil {
			_ = fr.Close()
			return err
		}
		_, copyErr := io.Copy(fw, fr)
		closeReadErr := fr.Close()
		closeWriteErr := fw.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeReadErr != nil {
			return closeReadErr
		}
		if closeWriteErr != nil {
			return closeWriteErr
		}
		if options.PreserveOwner && hasOwnership {
			if err := applyArchiveOwnership(filePath, info.Mode(), ownership); err != nil {
				return fmt.Errorf("restore archive ownership for %s: %w", fileName, err)
			}
		}
		if err := f.Fs.Chmod(filePath, info.Mode().Perm()); err != nil {
			return fmt.Errorf("restore archive mode for %s: %w", fileName, err)
		}
		_ = os.Chtimes(filePath, info.ModTime(), info.ModTime())
		return nil
	}
	if err := extractor.Extract(ctx, input, nil, handler); err != nil {
		return extractionStarted, err
	}
	for _, link := range hardlinks {
		if err := ensureArchiveParent(dst, link.path); err != nil {
			return extractionStarted, err
		}
		if err := validateArchiveHardlinkTarget(dst, link.target); err != nil {
			return extractionStarted, err
		}
		if err := os.RemoveAll(link.path); err != nil {
			return extractionStarted, err
		}
		if err := os.Link(link.target, link.path); err != nil {
			return extractionStarted, fmt.Errorf("restore archive hardlink %s: %w", link.path, err)
		}
		if options.PreserveOwner && link.hasOwnership {
			if err := applyArchiveOwnership(link.path, link.mode, link.ownership); err != nil {
				return extractionStarted, fmt.Errorf("restore archive ownership for %s: %w", link.path, err)
			}
		}
		if err := f.Fs.Chmod(link.path, link.mode.Perm()); err != nil {
			return extractionStarted, fmt.Errorf("restore archive mode for %s: %w", link.path, err)
		}
		_ = os.Chtimes(link.path, link.modTime, link.modTime)
	}
	for i := len(dirs) - 1; i >= 0; i-- {
		if options.PreserveOwner && dirs[i].hasOwnership {
			if err := applyArchiveOwnership(dirs[i].path, dirs[i].mode, dirs[i].ownership); err != nil {
				return extractionStarted, fmt.Errorf("restore archive ownership for %s: %w", dirs[i].path, err)
			}
		}
		if err := f.Fs.Chmod(dirs[i].path, dirs[i].mode.Perm()); err != nil {
			return extractionStarted, fmt.Errorf("restore archive mode for %s: %w", dirs[i].path, err)
		}
		_ = os.Chtimes(dirs[i].path, dirs[i].modTime, dirs[i].modTime)
	}
	return extractionStarted, nil
}

func (f FileOp) decompressWithSDKState(ctx context.Context, srcFile string, dst string, cType CompressType, secret string, options DecompressOptions) (bool, error) {
	input, err := f.Fs.Open(srcFile)
	if err != nil {
		return false, err
	}
	var extractor archiver.Extractor = getFormat(cType)
	if cType == X7z {
		extractor = archiver.SevenZip{Password: secret}
	}
	extractionStarted, extractErr := f.extractArchiveWithSDK(ctx, input, dst, extractor, options)
	closeErr := input.Close()
	if cType == Gz {
		if extractErr == nil && extractionStarted {
			return true, closeErr
		}
		if extractionStarted {
			return true, extractErr
		}
		return false, f.DecompressGzFile(ctx, srcFile, dst)
	}
	if extractErr != nil {
		return extractionStarted, extractErr
	}
	if closeErr != nil {
		return extractionStarted, closeErr
	}
	return extractionStarted, nil
}

func (f FileOp) decompressWithSDK(ctx context.Context, srcFile string, dst string, cType CompressType, secret string, options DecompressOptions) error {
	_, err := f.decompressWithSDKState(ctx, srcFile, dst, cType, secret, options)
	return err
}

func resetArchiveFallbackDestination(dst string) error {
	info, err := os.Lstat(dst)
	if os.IsNotExist(err) {
		return os.MkdirAll(dst, constant.DirPerm)
	}
	if err != nil {
		return err
	}
	if info.Mode()&fs.ModeSymlink != 0 || !info.IsDir() {
		return fmt.Errorf("archive fallback destination is not a directory: %s", dst)
	}
	mode := info.Mode().Perm()
	ownership, hasOwnership := getFileOwnership(info)
	if err := os.RemoveAll(dst); err != nil {
		return err
	}
	if err := os.MkdirAll(dst, mode); err != nil {
		return err
	}
	if hasOwnership {
		if err := os.Chown(dst, ownership.uid, ownership.gid); err != nil {
			return err
		}
	}
	return os.Chmod(dst, mode)
}

func (f FileOp) decompressSevenZipWithFallback(ctx context.Context, srcFile, dst, secret string, options DecompressOptions) error {
	return f.decompressSevenZipWithFallbackUsing(ctx, srcFile, dst, secret, options, f.decompressWithSDKState)
}

func (f FileOp) decompressSevenZipWithFallbackUsing(
	ctx context.Context,
	srcFile, dst, secret string,
	options DecompressOptions,
	sdkExtract func(context.Context, string, string, CompressType, string, DecompressOptions) (bool, error),
) error {
	extractionStarted, sdkErr := sdkExtract(ctx, srcFile, dst, X7z, secret, options)
	if sdkErr == nil {
		return nil
	}
	if secret != "" {
		return sdkErr
	}
	if extractionStarted && !options.AllowCLIReextract {
		return sdkErr
	}
	if extractionStarted {
		if err := resetArchiveFallbackDestination(dst); err != nil {
			return fmt.Errorf("reset 7z CLI fallback destination: %w", err)
		}
	}

	shellArchiver, err := NewExtractShellArchiver(X7z)
	if err != nil {
		return err
	}
	if global.LOG != nil {
		global.LOG.Warnf("7z SDK decompression failed, falling back to CLI: %v", sdkErr)
	}
	if err := shellArchiver.Extract(ctx, srcFile, dst, secret); err != nil {
		return fmt.Errorf("7z SDK decompression failed: %v; CLI fallback failed: %w", sdkErr, err)
	}
	return nil
}

type ownershipPreservingShellExtractor interface {
	ExtractWithOptions(ctx context.Context, filePath, dstDir, secret string, preserveOwner bool) error
}

func extractWithShellOptions(ctx context.Context, shellArchiver ShellArchiver, srcFile, dst, secret string, options DecompressOptions) error {
	if options.PreserveOwner {
		if extractor, ok := shellArchiver.(ownershipPreservingShellExtractor); ok {
			return extractor.ExtractWithOptions(ctx, srcFile, dst, secret, true)
		}
	}
	return shellArchiver.Extract(ctx, srcFile, dst, secret)
}

func (f FileOp) Decompress(ctx context.Context, srcFile string, dst string, cType CompressType, secret string) error {
	return f.DecompressWithOptions(ctx, srcFile, dst, cType, secret, DecompressOptions{})
}

func (f FileOp) DecompressWithOptions(ctx context.Context, srcFile string, dst string, cType CompressType, secret string, options DecompressOptions) error {
	if cType == X7z && options.PreserveOwner {
		return f.decompressSevenZipWithFallback(ctx, srcFile, dst, secret, options)
	}

	var shellErr error
	useShell := cType == Rar || cType == Zip || cType == Tar || cType == TarGz ||
		!options.PreserveOwner && cType == X7z
	if useShell {
		shellArchiver, err := NewExtractShellArchiver(cType)
		if !f.Stat(dst) {
			_ = f.CreateDir(dst, 0755)
		}
		if err == nil {
			if err = extractWithShellOptions(ctx, shellArchiver, srcFile, dst, secret, options); err == nil {
				return nil
			}
			shellErr = err
			if cType == TarGz {
				if strings.Contains(err.Error(), "bad decrypt") {
					return buserr.New("ErrBadDecrypt")
				}
				if retryErr := extractWithShellOptions(ctx, shellArchiver, srcFile, dst, "-", options); retryErr == nil {
					return nil
				} else if strings.Contains(retryErr.Error(), "bad decrypt") {
					return buserr.New("ErrBadDecrypt")
				} else {
					shellErr = retryErr
				}
			}
		} else {
			if cType == Rar || cType == X7z {
				return err
			}
		}
	}
	if shellErr != nil && global.LOG != nil {
		global.LOG.Warnf("shell decompression for %s failed, falling back to SDK: %v", cType, shellErr)
	}
	if shellErr != nil && options.AllowCLIReextract {
		if err := resetArchiveFallbackDestination(dst); err != nil {
			return fmt.Errorf("reset %s SDK fallback destination: %w", cType, err)
		}
	}
	return f.decompressWithSDK(ctx, srcFile, dst, cType, secret, options)
}

func ZipFile(ctx context.Context, files []archiver.File, dst afero.File, progress func(current, total int, message string)) error {
	zw := zip.NewWriter(dst)
	defer zw.Close()

	total := len(files)
	for i, file := range files {
		if ctx != nil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
		}
		hdr, err := zip.FileInfoHeader(file)
		if err != nil {
			return err
		}
		hdr.Method = zip.Deflate
		hdr.Name = file.NameInArchive
		if ownership, ok := getFileOwnership(file.FileInfo); ok {
			hdr.Extra = appendZipOwnershipExtra(hdr.Extra, ownership)
		}
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
			_, err = io.Copy(w, newContextReader(ctx, fileReader))
			fileReader.Close()
			if err != nil {
				return err
			}
		}
		if progress != nil {
			progress(i+1, total, file.NameInArchive)
		}
	}
	return nil
}

type contextReader struct {
	ctx context.Context
	r   io.Reader
}

func newContextReader(ctx context.Context, r io.Reader) io.Reader {
	if ctx == nil {
		return r
	}
	return &contextReader{ctx: ctx, r: r}
}

func (r *contextReader) Read(p []byte) (int, error) {
	select {
	case <-r.ctx.Done():
		return 0, r.ctx.Err()
	default:
		return r.r.Read(p)
	}
}

func (f FileOp) DecompressGzFile(ctx context.Context, srcFile, dst string) error {
	var archiveModTime time.Time
	if st, err := f.Fs.Stat(srcFile); err == nil {
		archiveModTime = st.ModTime()
	}

	in, err := f.Fs.Open(srcFile)
	if err != nil {
		return fmt.Errorf("open source file failed: %w", err)
	}
	defer in.Close()

	gr, err := gzip.NewReader(&contextReader{ctx: ctx, r: in})
	if err != nil {
		return fmt.Errorf("gzip reader creation failed: %w", err)
	}
	defer gr.Close()

	outName := ""
	if gr.Name != "" {
		outName = filepath.Base(gr.Name)
	}
	if outName == "" || outName == "." {
		outName = strings.TrimSuffix(filepath.Base(srcFile), ".gz")
	}
	outPath := filepath.Join(dst, outName)
	parentDir := filepath.Dir(outPath)
	if !f.Stat(parentDir) {
		if err := f.Fs.MkdirAll(parentDir, 0755); err != nil {
			return err
		}
	}

	fw, err := f.Fs.OpenFile(outPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("create output file failed: %w", err)
	}
	defer fw.Close()

	if _, err := io.Copy(fw, gr); err != nil {
		return fmt.Errorf("copy content failed: %w", err)
	}

	if !archiveModTime.IsZero() {
		_ = os.Chtimes(outPath, archiveModTime, archiveModTime)
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
	exMap := make(map[string]struct{})
	excludeArgs := []string{}
	excludes := strings.Split(exclusionRules, ",")
	for _, exclude := range excludes {
		if len(exclude) == 0 {
			continue
		}
		if strings.HasPrefix(exclude, "/") {
			exclude, _ = filepath.Rel(src, exclude)
		}
		if _, ok := exMap[exclude]; ok {
			continue
		}
		excludeArgs = append(excludeArgs, "--exclude", exclude)
		exMap[exclude] = struct{}{}
	}

	tarArgs := append([]string{}, excludeArgs...)
	if len(secret) != 0 {
		cmdMgr := cmd.NewCommandMgr(cmd.WithWorkDir(workdir), cmd.WithIgnoreExist1())
		return runTarGzEncryptToFile(cmdMgr, dst, secret, append(tarArgs, srcItem)...)
	} else {
		cmdMgr := cmd.NewCommandMgr(cmd.WithWorkDir(workdir), cmd.WithIgnoreExist1())
		return runTarGzToFile(cmdMgr, dst, append(tarArgs, srcItem)...)
	}
}

func (f FileOp) TarGzFilesWithCompressPro(list []string, dst, secret string) error {
	if !f.Stat(path.Dir(dst)) {
		if err := f.Fs.MkdirAll(path.Dir(dst), constant.FilePerm); err != nil {
			return err
		}
	}

	var tarArgs []string
	for _, item := range list {
		tarArgs = append(tarArgs, "-C", path.Dir(item), path.Base(item))
	}
	if len(secret) != 0 {
		cmdMgr := cmd.NewCommandMgr(cmd.WithIgnoreExist1())
		return runTarGzEncryptToFile(cmdMgr, dst, secret, tarArgs...)
	} else {
		cmdMgr := cmd.NewCommandMgr(cmd.WithIgnoreExist1())
		return runTarGzToFile(cmdMgr, dst, tarArgs...)
	}
}

func (f FileOp) TarGzExtractPro(src, dst string, secret string) error {
	if _, err := os.Stat(dst); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(dst, os.ModePerm); err != nil {
			return err
		}
	}

	if len(secret) != 0 {
		cmdMgr := cmd.NewCommandMgr(cmd.WithWorkDir(dst), cmd.WithIgnoreExist1())
		return runTarGzDecryptToDir(cmdMgr, src, dst, secret, true)
	} else {
		cmdMgr := cmd.NewCommandMgr(cmd.WithWorkDir(dst), cmd.WithIgnoreExist1())
		return runTarGzExtractToDir(cmdMgr, src, dst)
	}
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

func OpensslEncrypt(filePath, secret string) error {
	tmpName := path.Join(path.Dir(filePath), "tmp_"+path.Base(filePath))
	if err := cmd.NewCommandMgr(cmd.WithEnv("MY_PASS="+secret)).Run("openssl", "enc", "-aes-256-cbc", "-salt", "-pass", "env:MY_PASS", "-in", filePath, "-out", tmpName); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, filePath)
}

func OpensslDecrypt(filePath, secret string) error {
	tmpName := path.Join(path.Dir(filePath), "tmp_"+path.Base(filePath))
	if err := cmd.NewCommandMgr(cmd.WithEnv("MY_PASS="+secret)).Run("openssl", "enc", "-aes-256-cbc", "-d", "-salt", "-pass", "env:MY_PASS", "-in", filePath, "-out", tmpName); err != nil {
		if strings.Contains(err.Error(), "bad decrypt") || strings.Contains(err.Error(), "bad magic number") {
			return buserr.New("ErrBadDecrypt")
		}
		return err
	}
	return nil
}
