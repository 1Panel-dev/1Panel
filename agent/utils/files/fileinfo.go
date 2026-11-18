package files

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"

	"github.com/spf13/afero"
)

var binaryPreviewMimeTypes = map[string]struct{}{
	"application/pdf":               {},
	"application/zip":               {},
	"application/gzip":              {},
	"application/x-gzip":            {},
	"application/x-7z-compressed":   {},
	"application/x-rar-compressed":  {},
	"application/x-bzip2":           {},
	"application/x-xz":              {},
	"application/x-tar":             {},
	"application/java-archive":      {},
	"application/msword":            {},
	"application/vnd.ms-excel":      {},
	"application/vnd.ms-powerpoint": {},
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   {},
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         {},
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": {},
}

var binaryPreviewMimePrefixes = []string{
	"image/",
	"audio/",
	"video/",
	"application/vnd.ms-",
	"application/vnd.openxmlformats-officedocument.",
	"application/vnd.oasis.opendocument.",
}

var binaryPreviewExtensions = map[string]struct{}{
	".pdf":  {},
	".zip":  {},
	".gz":   {},
	".bz2":  {},
	".xz":   {},
	".tar":  {},
	".tgz":  {},
	".rar":  {},
	".7z":   {},
	".war":  {},
	".doc":  {},
	".docx": {},
	".xls":  {},
	".xlsx": {},
	".ppt":  {},
	".pptx": {},
	".jpg":  {},
	".jpeg": {},
	".png":  {},
	".bmp":  {},
	".gif":  {},
	".tiff": {},
	".ico":  {},
	".webp": {},
	".svg":  {},
	".mp3":  {},
	".wav":  {},
	".wma":  {},
	".ape":  {},
	".acc":  {},
	".ogg":  {},
	".flac": {},
	".mp4":  {},
	".webm": {},
	".mov":  {},
	".wmv":  {},
	".mkv":  {},
	".avi":  {},
	".flv":  {},
}

type FileInfo struct {
	Fs           afero.Fs    `json:"-"`
	Path         string      `json:"path"`
	Name         string      `json:"name"`
	User         string      `json:"user"`
	Group        string      `json:"group"`
	Uid          string      `json:"uid"`
	Gid          string      `json:"gid"`
	Extension    string      `json:"extension"`
	Content      string      `json:"content"`
	Size         int64       `json:"size"`
	IsDir        bool        `json:"isDir"`
	IsSymlink    bool        `json:"isSymlink"`
	IsHidden     bool        `json:"isHidden"`
	LinkPath     string      `json:"linkPath"`
	Type         string      `json:"type"`
	Mode         string      `json:"mode"`
	MimeType     string      `json:"mimeType"`
	UpdateTime   time.Time   `json:"updateTime"`
	ModTime      time.Time   `json:"modTime"`
	FileMode     os.FileMode `json:"-"`
	Items        []*FileInfo `json:"items"`
	ItemTotal    int         `json:"itemTotal"`
	FavoriteID   uint        `json:"favoriteID"`
	ShareCode    string      `json:"shareCode"`
	IsDetail     bool        `json:"isDetail"`
	IsAppendOnly bool        `json:"isAppendOnly"`
	IsImmutable  bool        `json:"isImmutable"`
}

type FileOption struct {
	Path       string `json:"path"`
	Search     string `json:"search"`
	ContainSub bool   `json:"containSub"`
	Expand     bool   `json:"expand"`
	Dir        bool   `json:"dir"`
	ShowHidden bool   `json:"showHidden"`
	Page       int    `json:"page"`
	PageSize   int    `json:"pageSize"`
	SortBy     string `json:"sortBy"`
	SortOrder  string `json:"sortOrder"`
	IsDetail   bool   `json:"isDetail"`
}

type FileSearchInfo struct {
	Path string `json:"path"`
	fs.FileInfo
}

func NewFileInfo(op FileOption) (*FileInfo, error) {
	return NewFileInfoWithContext(context.Background(), op)
}

func NewFileInfoWithContext(ctx context.Context, op FileOption) (*FileInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var appFs = afero.NewOsFs()

	info, err := appFs.Stat(op.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, buserr.New("ErrLinkPathNotFound")
		}
		return nil, err
	}

	file := &FileInfo{
		Fs:        appFs,
		Path:      op.Path,
		Name:      info.Name(),
		IsDir:     info.IsDir(),
		FileMode:  info.Mode(),
		ModTime:   info.ModTime(),
		Size:      info.Size(),
		IsSymlink: IsSymlink(info.Mode()),
		Extension: filepath.Ext(info.Name()),
		IsHidden:  IsHidden(op.Path),
		Mode:      fmt.Sprintf("%04o", info.Mode().Perm()),
		User:      GetUsername(info.Sys().(*syscall.Stat_t).Uid),
		Uid:       strconv.FormatUint(uint64(info.Sys().(*syscall.Stat_t).Uid), 10),
		Gid:       strconv.FormatUint(uint64(info.Sys().(*syscall.Stat_t).Gid), 10),
		Group:     GetGroup(info.Sys().(*syscall.Stat_t).Gid),
		IsDetail:  op.IsDetail,
	}
	if info.Mode().IsRegular() {
		file.MimeType = GetMimeType(op.Path)
	}
	file.IsAppendOnly, file.IsImmutable = getFileAttributes(ctx, op.Path)

	if file.IsSymlink {
		linkPath := GetSymlink(op.Path)
		if !filepath.IsAbs(linkPath) {
			dir := filepath.Dir(op.Path)
			var err error
			linkPath, err = filepath.Abs(filepath.Join(dir, linkPath))
			if err != nil {
				return nil, err
			}
		}
		file.LinkPath = linkPath
		targetInfo, err := appFs.Stat(linkPath)
		if err != nil {
			file.IsDir = false
			file.Mode = "-"
			file.User = "-"
			file.Group = "-"
		} else {
			file.IsDir = targetInfo.IsDir()
		}
		file.Extension = filepath.Ext(file.LinkPath)
	}
	if op.Expand {
		if err := handleExpansion(ctx, file, op); err != nil {
			return nil, err
		}
	}
	if !op.Expand || !file.IsDir {
		if err := file.loadFavorites(ctx); err != nil {
			return nil, err
		}
	}
	return file, nil
}

func getFileAttributes(ctx context.Context, filePath string) (bool, bool) {
	output, err := exec.CommandContext(ctx, "lsattr", "-d", "--", filePath).Output()
	if err != nil {
		return false, false
	}
	fields := strings.Fields(string(output))
	if len(fields) == 0 {
		return false, false
	}
	return strings.Contains(fields[0], "a"), strings.Contains(fields[0], "i")
}

func handleExpansion(ctx context.Context, file *FileInfo, op FileOption) error {
	if file.IsDir {
		return file.listChildrenWithContext(ctx, op)
	}

	if !file.IsDetail {
		return file.getContent()
	}

	return nil
}

func (f *FileInfo) search(ctx context.Context, search string, count int) (files []FileSearchInfo, total int, err error) {
	searchCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	cmd := exec.CommandContext(searchCtx, "find", f.Path, "-name", fmt.Sprintf("*%s*", search), "-print0")
	output, err := cmd.StdoutPipe()
	if err != nil {
		return nil, 0, err
	}
	if err := cmd.Start(); err != nil {
		return nil, 0, err
	}
	reader := bufio.NewReader(output)
	for {
		if err = ctx.Err(); err != nil {
			break
		}
		var entry string
		entry, err = reader.ReadString(0)
		if err != nil {
			break
		}
		entry = strings.TrimSuffix(entry, "\x00")
		info, statErr := os.Stat(entry)
		if statErr != nil {
			continue
		}
		total++
		if total <= count {
			files = append(files, FileSearchInfo{Path: entry, FileInfo: info})
		}
	}
	if err != nil && err != io.EOF {
		cancel()
	}
	_ = cmd.Wait()
	if contextErr := ctx.Err(); contextErr != nil {
		return nil, 0, contextErr
	}
	if err == io.EOF {
		err = nil
	}
	return
}

func sortFileList(list []FileSearchInfo, sortBy, sortOrder string) {
	_ = sortFileListWithContext(context.Background(), list, sortBy, sortOrder)
}

func sortFileListWithContext(ctx context.Context, list []FileSearchInfo, sortBy, sortOrder string) error {
	var less func(i, j int) bool
	switch sortBy {
	case "name":
		less = func(i, j int) bool { return list[i].Name() < list[j].Name() }
	case "size":
		less = func(i, j int) bool { return list[i].Size() < list[j].Size() }
	case "modTime":
		less = func(i, j int) bool { return list[i].ModTime().Before(list[j].ModTime()) }
	default:
		return ctx.Err()
	}
	return sortSliceWithContext(ctx, list, func(i, j int) bool {
		if sortOrder != "ascending" {
			i, j = j, i
		}
		return less(i, j)
	})
}

func sortSliceWithContext(ctx context.Context, items interface{}, less func(i, j int) bool) (err error) {
	if err := ctx.Err(); err != nil {
		return err
	}
	if ctx.Done() == nil {
		sort.Slice(items, less)
		return nil
	}
	canceled := &struct{ marker byte }{}
	defer func() {
		if recovered := recover(); recovered != nil {
			if recovered != canceled {
				panic(recovered)
			}
			err = ctx.Err()
		}
	}()
	comparisons := 0
	sort.Slice(items, func(i, j int) bool {
		if comparisons%256 == 0 && ctx.Err() != nil {
			panic(canceled)
		}
		comparisons++
		return less(i, j)
	})
	return ctx.Err()
}

func (f *FileInfo) listChildren(option FileOption) error {
	return f.listChildrenWithContext(context.Background(), option)
}

func (f *FileInfo) listChildrenWithContext(ctx context.Context, option FileOption) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	maxInt := int(^uint(0) >> 1)
	if option.Page < 1 || option.PageSize < 1 || option.Page > maxInt/option.PageSize {
		return fmt.Errorf("invalid file pagination")
	}
	var (
		entries []FileSearchInfo
		total   int
		err     error
	)
	if option.Search != "" && option.ContainSub {
		entries, total, err = f.search(ctx, option.Search, option.Page*option.PageSize)
	} else {
		entries, err = f.getFiles(ctx, option)
	}
	if err != nil {
		return err
	}

	candidates := entries[:0]
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return err
		}
		if shouldSkipFile(entry, option) {
			continue
		}
		name, _ := f.getFilePathAndName(option, entry)
		if !option.ShowHidden && IsHidden(name) {
			continue
		}
		candidates = append(candidates, entry)
	}
	f.ItemTotal = len(candidates)
	if option.ContainSub {
		f.ItemTotal = total
	}

	page := candidates
	if !option.ContainSub || option.Search != "" {
		start := (option.Page - 1) * option.PageSize
		end := option.Page * option.PageSize
		if start >= len(candidates) {
			page = nil
		} else {
			if end > len(candidates) {
				end = len(candidates)
			}
			page = candidates[start:end]
		}
	}
	f.Items, err = f.processFiles(ctx, page, option)
	if err != nil {
		return err
	}
	return f.loadFavorites(ctx)
}

func readDirWithContext(ctx context.Context, dirPath string) ([]os.DirEntry, error) {
	dir, err := os.Open(dirPath)
	if err != nil {
		return nil, err
	}
	defer dir.Close()
	var entries []os.DirEntry
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		batch, err := dir.ReadDir(256)
		entries = append(entries, batch...)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}
	if err := sortSliceWithContext(ctx, entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() }); err != nil {
		return nil, err
	}
	return entries, nil
}

func (f *FileInfo) getFiles(ctx context.Context, option FileOption) ([]FileSearchInfo, error) {
	entries, err := readDirWithContext(ctx, f.Path)
	if err != nil {
		return nil, err
	}
	var dirs, fileList []FileSearchInfo
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		item := FileSearchInfo{Path: f.Path, FileInfo: info}
		if entry.IsDir() {
			dirs = append(dirs, item)
		} else {
			fileList = append(fileList, item)
		}
	}
	if err := sortFileListWithContext(ctx, dirs, option.SortBy, option.SortOrder); err != nil {
		return nil, err
	}
	if err := sortFileListWithContext(ctx, fileList, option.SortBy, option.SortOrder); err != nil {
		return nil, err
	}
	return append(dirs, fileList...), ctx.Err()
}

func (f *FileInfo) processFiles(ctx context.Context, entries []FileSearchInfo, option FileOption) ([]*FileInfo, error) {
	var items []*FileInfo
	users := make(map[uint32]string)
	groups := make(map[uint32]string)
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		name, filePath := f.getFilePathAndName(option, entry)
		stat := entry.Sys().(*syscall.Stat_t)
		username, ok := users[stat.Uid]
		if !ok {
			username = GetUsername(stat.Uid)
			users[stat.Uid] = username
		}
		group, ok := groups[stat.Gid]
		if !ok {
			group = GetGroup(stat.Gid)
			groups[stat.Gid] = group
		}
		isSymlink, isInvalidLink := f.checkSymlink(entry)
		file := &FileInfo{
			Fs: f.Fs, Name: name, Path: filePath,
			Size: entry.Size(), ModTime: entry.ModTime(), FileMode: entry.Mode(),
			IsDir: entry.IsDir(), IsSymlink: isSymlink, IsHidden: IsHidden(filePath),
			Extension: filepath.Ext(name), Mode: fmt.Sprintf("%04o", entry.Mode().Perm()),
			User: username, Group: group,
			Uid: strconv.FormatUint(uint64(stat.Uid), 10),
			Gid: strconv.FormatUint(uint64(stat.Gid), 10),
		}
		mimeInfo := entry.FileInfo
		if isSymlink {
			linkPath := GetSymlink(filePath)
			if !filepath.IsAbs(linkPath) {
				resolved, err := filepath.Abs(filepath.Join(filepath.Dir(filePath), linkPath))
				if err != nil {
					return nil, err
				}
				linkPath = resolved
			}
			file.LinkPath = linkPath
			targetInfo, err := file.Fs.Stat(linkPath)
			mimeInfo = targetInfo
			if err != nil {
				file.IsDir = false
				file.Mode, file.User, file.Group = "-", "-", "-"
			} else {
				file.IsDir = targetInfo.IsDir()
			}
			file.Extension = filepath.Ext(file.LinkPath)
		}
		if entry.Size() > 0 && mimeInfo != nil && mimeInfo.Mode().IsRegular() {
			file.MimeType = getMimeTypeFromFS(f.Fs, filePath)
		}
		if isInvalidLink {
			file.Type = "invalid_link"
		}
		items = append(items, file)
	}
	return items, ctx.Err()
}

func (f *FileInfo) loadFavorites(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	paths := make([]string, 0, len(f.Items)+1)
	paths = append(paths, f.Path)
	for _, item := range f.Items {
		paths = append(paths, item.Path)
	}
	favorites, err := repo.NewIFavoriteRepo().ListByPaths(ctx, paths)
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if err != nil {
		return nil
	}
	ids := make(map[string]uint, len(favorites))
	for _, favorite := range favorites {
		if id, ok := ids[favorite.Path]; !ok || favorite.ID < id {
			ids[favorite.Path] = favorite.ID
		}
	}
	f.FavoriteID = ids[f.Path]
	for _, item := range f.Items {
		item.FavoriteID = ids[item.Path]
	}
	return nil
}

func shouldSkipFile(df FileSearchInfo, option FileOption) bool {
	if option.Dir && !df.IsDir() {
		return true
	}

	if option.Search != "" && !option.ContainSub {
		lowerName := strings.ToLower(df.Name())
		lowerSearch := strings.ToLower(option.Search)
		if !strings.Contains(lowerName, lowerSearch) {
			return true
		}
	}

	return false
}

func (f *FileInfo) getFilePathAndName(option FileOption, df FileSearchInfo) (string, string) {
	name := df.Name()
	fPath := path.Join(df.Path, df.Name())

	if option.Search != "" && option.ContainSub {
		fPath = df.Path
		name = strings.TrimPrefix(strings.TrimPrefix(fPath, f.Path), "/")
	}

	return name, fPath
}

func (f *FileInfo) checkSymlink(df FileSearchInfo) (bool, bool) {
	isSymlink := false
	isInvalidLink := false

	if IsSymlink(df.Mode()) {
		isSymlink = true
		info, err := f.Fs.Stat(path.Join(df.Path, df.Name()))
		if err == nil {
			df.FileInfo = info
		} else {
			isInvalidLink = true
		}
	}

	return isSymlink, isInvalidLink
}

func (f *FileInfo) getContent() error {
	if IsBlockDevice(f.FileMode) {
		return buserr.New("ErrFileCanNotRead")
	}
	if f.Size > 10*1024*1024 {
		return buserr.New("ErrFileToLarge")
	}
	if IsBinaryPreviewFile(f.MimeType, f.Extension) {
		return buserr.New("ErrFileCanNotRead")
	}
	afs := &afero.Afero{Fs: f.Fs}
	cByte, err := afs.ReadFile(f.Path)
	if err != nil {
		return nil
	}
	if len(cByte) > 0 && DetectBinary(cByte) {
		return buserr.New("ErrFileCanNotRead")
	}
	f.Content = string(cByte)
	return nil
}

func IsBinaryPreviewFile(mimeType, extension string) bool {
	mimeType = strings.ToLower(strings.TrimSpace(mimeType))
	extension = strings.ToLower(strings.TrimSpace(extension))

	if mimeType == "" || mimeType == "application/octet-stream" {
		_, ok := binaryPreviewExtensions[extension]
		return ok
	}

	if strings.HasPrefix(mimeType, "text/") {
		return false
	}

	for _, prefix := range binaryPreviewMimePrefixes {
		if strings.HasPrefix(mimeType, prefix) {
			return true
		}
	}

	_, ok := binaryPreviewMimeTypes[mimeType]
	if ok {
		return true
	}

	_, ok = binaryPreviewExtensions[extension]
	return ok
}

func DetectBinary(buf []byte) bool {
	mimeType := http.DetectContentType(buf)
	if IsBinaryPreviewFile(mimeType, "") {
		return true
	}
	if !strings.HasPrefix(mimeType, "text/") {
		whiteByte := 0
		n := min(1024, len(buf))
		for i := 0; i < n; i++ {
			if (buf[i] >= 0x20) || buf[i] == 9 || buf[i] == 10 || buf[i] == 13 {
				whiteByte++
			} else if buf[i] <= 6 || (buf[i] >= 14 && buf[i] <= 31) {
				return true
			}
		}
		return whiteByte < 1
	}
	return false
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

type CompressType string

const (
	Zip      CompressType = "zip"
	Gz       CompressType = "gz"
	Bz2      CompressType = "bz2"
	TarBz2   CompressType = "tar.bz2"
	Tar      CompressType = "tar"
	TarGz    CompressType = "tar.gz"
	Tgz      CompressType = "tgz"
	Xz       CompressType = "xz"
	TarXz    CompressType = "tar.xz"
	SdkZip   CompressType = "sdkZip"
	SdkTarGz CompressType = "sdkTarGz"
	Rar      CompressType = "rar"
	X7z      CompressType = "7z"
)
