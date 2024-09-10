package files

import (
	"bufio"
	"fmt"
	"github.com/1Panel-dev/1Panel/backend/app/repo"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
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

	"github.com/spf13/afero"
)

type FileInfo struct {
	Fs         afero.Fs    `json:"-"`
	Path       string      `json:"path"`
	Name       string      `json:"name"`
	User       string      `json:"user"`
	Group      string      `json:"group"`
	Uid        string      `json:"uid"`
	Gid        string      `json:"gid"`
	Extension  string      `json:"extension"`
	Content    string      `json:"content"`
	Size       int64       `json:"size"`
	IsDir      bool        `json:"isDir"`
	IsSymlink  bool        `json:"isSymlink"`
	IsHidden   bool        `json:"isHidden"`
	LinkPath   string      `json:"linkPath"`
	Type       string      `json:"type"`
	Mode       string      `json:"mode"`
	MimeType   string      `json:"mimeType"`
	UpdateTime time.Time   `json:"updateTime"`
	ModTime    time.Time   `json:"modTime"`
	FileMode   os.FileMode `json:"-"`
	Items      []*FileInfo `json:"items"`
	ItemTotal  int         `json:"itemTotal"`
	FavoriteID uint        `json:"favoriteID"`
	IsDetail   bool        `json:"isDetail"`
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
	var appFs = afero.NewOsFs()

	info, err := appFs.Stat(op.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, buserr.New(constant.ErrLinkPathNotFound)
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
		MimeType:  GetMimeType(op.Path),
		IsDetail:  op.IsDetail,
	}
	favoriteRepo := repo.NewIFavoriteRepo()
	favorite, _ := favoriteRepo.GetFirst(favoriteRepo.WithByPath(op.Path))
	if favorite.ID > 0 {
		file.FavoriteID = favorite.ID
	}

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
		if err := handleExpansion(file, op); err != nil {
			return nil, err
		}
	}
	return file, nil
}

func handleExpansion(file *FileInfo, op FileOption) error {
	if file.IsDir {
		return file.listChildren(op)
	}

	if !file.IsDetail {
		return file.getContent()
	}

	return nil
}

func (f *FileInfo) search(search string, count int) (files []FileSearchInfo, total int, err error) {
	cmd := exec.Command("find", f.Path, "-name", fmt.Sprintf("*%s*", search))
	output, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	if err = cmd.Start(); err != nil {
		return
	}
	defer func() {
		_ = cmd.Wait()
		_ = cmd.Process.Kill()
	}()

	scanner := bufio.NewScanner(output)
	for scanner.Scan() {
		line := scanner.Text()
		info, err := os.Stat(line)
		if err != nil {
			continue
		}
		total++
		if total > count {
			continue
		}
		files = append(files, FileSearchInfo{
			Path:     line,
			FileInfo: info,
		})
	}
	if err = scanner.Err(); err != nil {
		return
	}
	return
}

func sortFileList(list []FileSearchInfo, sortBy, sortOrder string) {
	switch sortBy {
	case "name":
		if sortOrder == "ascending" {
			sort.Slice(list, func(i, j int) bool {
				return list[i].Name() < list[j].Name()
			})
		} else {
			sort.Slice(list, func(i, j int) bool {
				return list[i].Name() > list[j].Name()
			})
		}
	case "size":
		if sortOrder == "ascending" {
			sort.Slice(list, func(i, j int) bool {
				return list[i].Size() < list[j].Size()
			})
		} else {
			sort.Slice(list, func(i, j int) bool {
				return list[i].Size() > list[j].Size()
			})
		}
	case "modTime":
		if sortOrder == "ascending" {
			sort.Slice(list, func(i, j int) bool {
				return list[i].ModTime().Before(list[j].ModTime())
			})
		} else {
			sort.Slice(list, func(i, j int) bool {
				return list[i].ModTime().After(list[j].ModTime())
			})
		}
	}
}

func (f *FileInfo) listChildren(option FileOption) error {
	afs := &afero.Afero{Fs: f.Fs}
	var (
		files []FileSearchInfo
		err   error
		total int
	)

	if option.Search != "" && option.ContainSub {
		files, total, err = f.search(option.Search, option.Page*option.PageSize)
		if err != nil {
			return err
		}
	} else {
		files, err = f.getFiles(afs, option)
		if err != nil {
			return err
		}
	}

	items, err := f.processFiles(files, option)
	if err != nil {
		return err
	}

	if option.ContainSub {
		f.ItemTotal = total
	}
	start := (option.Page - 1) * option.PageSize
	end := option.PageSize + start
	var result []*FileInfo
	if start < 0 || start > f.ItemTotal || end < 0 || start > end {
		result = items
	} else {
		if end > f.ItemTotal {
			result = items[start:]
		} else {
			result = items[start:end]
		}
	}

	f.Items = result
	return nil
}

func (f *FileInfo) getFiles(afs *afero.Afero, option FileOption) ([]FileSearchInfo, error) {
	dirFiles, err := afs.ReadDir(f.Path)
	if err != nil {
		return nil, err
	}

	var (
		dirs     []FileSearchInfo
		fileList []FileSearchInfo
	)

	for _, file := range dirFiles {
		info := FileSearchInfo{
			Path:     f.Path,
			FileInfo: file,
		}
		if file.IsDir() {
			dirs = append(dirs, info)
		} else {
			fileList = append(fileList, info)
		}
	}

	sortFileList(dirs, option.SortBy, option.SortOrder)
	sortFileList(fileList, option.SortBy, option.SortOrder)

	return append(dirs, fileList...), nil
}

func (f *FileInfo) processFiles(files []FileSearchInfo, option FileOption) ([]*FileInfo, error) {
	var items []*FileInfo

	for _, df := range files {
		if shouldSkipFile(df, option) {
			continue
		}

		name, fPath := f.getFilePathAndName(option, df)

		if !option.ShowHidden && IsHidden(name) {
			continue
		}
		f.ItemTotal++

		isSymlink, isInvalidLink := f.checkSymlink(df)

		file := &FileInfo{
			Fs:        f.Fs,
			Name:      name,
			Size:      df.Size(),
			ModTime:   df.ModTime(),
			FileMode:  df.Mode(),
			IsDir:     df.IsDir(),
			IsSymlink: isSymlink,
			IsHidden:  IsHidden(fPath),
			Extension: filepath.Ext(name),
			Path:      fPath,
			Mode:      fmt.Sprintf("%04o", df.Mode().Perm()),
			User:      GetUsername(df.Sys().(*syscall.Stat_t).Uid),
			Group:     GetGroup(df.Sys().(*syscall.Stat_t).Gid),
			Uid:       strconv.FormatUint(uint64(df.Sys().(*syscall.Stat_t).Uid), 10),
			Gid:       strconv.FormatUint(uint64(df.Sys().(*syscall.Stat_t).Gid), 10),
		}
		favoriteRepo := repo.NewIFavoriteRepo()
		favorite, _ := favoriteRepo.GetFirst(favoriteRepo.WithByPath(fPath))
		if favorite.ID > 0 {
			file.FavoriteID = favorite.ID
		}
		if isSymlink {
			linkPath := GetSymlink(fPath)
			if !filepath.IsAbs(linkPath) {
				dir := filepath.Dir(fPath)
				var err error
				linkPath, err = filepath.Abs(filepath.Join(dir, linkPath))
				if err != nil {
					return nil, err
				}
			}
			file.LinkPath = linkPath
			targetInfo, err := file.Fs.Stat(linkPath)
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
		if df.Size() > 0 {
			file.MimeType = GetMimeType(fPath)
		}
		if isInvalidLink {
			file.Type = "invalid_link"
		}
		items = append(items, file)
	}

	return items, nil
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
		return buserr.New(constant.ErrFileCanNotRead)
	}
	if f.Size > 10*1024*1024 {
		return buserr.New("ErrFileToLarge")
	}
	afs := &afero.Afero{Fs: f.Fs}
	cByte, err := afs.ReadFile(f.Path)
	if err != nil {
		return nil
	}
	if len(cByte) > 0 && DetectBinary(cByte) {
		return buserr.New(constant.ErrFileCanNotRead)
	}
	f.Content = string(cByte)
	return nil
}

func DetectBinary(buf []byte) bool {
	mimeType := http.DetectContentType(buf)
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
	Tar      CompressType = "tar"
	TarGz    CompressType = "tar.gz"
	Xz       CompressType = "xz"
	SdkZip   CompressType = "sdkZip"
	SdkTarGz CompressType = "sdkTarGz"
)
