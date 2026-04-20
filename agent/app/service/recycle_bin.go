package service

import (
	"bytes"
	"fmt"
	"math"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/dto/response"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/1Panel-dev/1Panel/agent/utils/re"
	"github.com/shirou/gopsutil/v4/disk"
	"gopkg.in/ini.v1"
)

const (
	recycleBinClashDir    = ".1panel_clash"
	recycleBinFilesSubdir = "files"
	recycleBinInfoSubdir  = "info"
	trashInfoSuffix       = ".trashinfo"
	trashInfoSection      = "Trash Info"
	trashInfoTimeLayout   = "2006-01-02T15:04:05"
)

type RecycleBinService struct {
}

type IRecycleBinService interface {
	Page(search dto.PageInfo) (int64, []response.RecycleBinDTO, error)
	Create(create request.RecycleBinCreate) error
	Reduce(reduce request.RecycleBinReduce) error
	Clear() error
}

func NewIRecycleBinService() IRecycleBinService {
	return &RecycleBinService{}
}

func (r RecycleBinService) Page(search dto.PageInfo) (int64, []response.RecycleBinDTO, error) {
	var result []response.RecycleBinDTO
	partitions, err := disk.Partitions(false)
	if err != nil {
		return 0, nil, err
	}
	op := files.NewFileOp()
	for _, p := range partitions {
		clashRoot := path.Join(p.Mountpoint, recycleBinClashDir)
		if !op.Stat(clashRoot) {
			continue
		}
		result = append(result, collectTrashEntries(clashRoot)...)
		result = append(result, collectLegacyEntries(clashRoot)...)
	}
	startIndex := (search.Page - 1) * search.PageSize
	endIndex := startIndex + search.PageSize

	if startIndex > len(result) {
		return int64(len(result)), result, nil
	}
	if endIndex > len(result) {
		endIndex = len(result)
	}
	return int64(len(result)), result[startIndex:endIndex], nil
}

func (r RecycleBinService) Create(create request.RecycleBinCreate) error {
	if files.IsProtected(create.SourcePath) {
		return buserr.New("ErrPathNotDelete")
	}
	op := files.NewFileOp()
	if !op.Stat(create.SourcePath) {
		return buserr.New("ErrLinkPathNotFound")
	}
	clashRoot, err := getClashDir(create.SourcePath)
	if err != nil {
		return err
	}
	filesDir, infoDir, err := ensureTrashDirs(clashRoot)
	if err != nil {
		return err
	}

	deleteTime := time.Now()
	openFile, err := op.OpenFile(create.SourcePath)
	if err != nil {
		return err
	}
	fileInfo, err := openFile.Stat()
	_ = openFile.Close()
	if err != nil {
		return err
	}
	size := int64(0)
	if fileInfo.IsDir() {
		sizeF, err := op.GetDirSize(create.SourcePath)
		if err != nil {
			return err
		}
		size = int64(sizeF)
	} else {
		size = fileInfo.Size()
	}

	rName := allocateTrashEntryName(filesDir, infoDir, create.SourcePath)
	info := trashInfo{
		Path:         create.SourcePath,
		DeletionDate: deleteTime,
		Size:         size,
		IsDir:        fileInfo.IsDir(),
	}
	infoPath := path.Join(infoDir, rName+trashInfoSuffix)
	if err := writeTrashInfo(infoPath, info); err != nil {
		return err
	}
	targetPath := path.Join(filesDir, rName)
	if err := op.Mv(create.SourcePath, targetPath); err != nil {
		if rmErr := os.Remove(infoPath); rmErr != nil && !os.IsNotExist(rmErr) {
			global.LOG.Warnf("rollback trashinfo failed for %s: %v", infoPath, rmErr)
		}
		return err
	}
	return nil
}

func (r RecycleBinService) Reduce(reduce request.RecycleBinReduce) error {
	filePath := path.Join(reduce.From, reduce.RName)
	op := files.NewFileOp()
	if !op.Stat(filePath) {
		return buserr.New("ErrLinkPathNotFound")
	}
	recycleBinDTO, err := loadRecycleBinDTO(reduce.From, reduce.RName)
	if err != nil {
		return err
	}
	if recycleBinDTO.SourcePath == "" {
		return buserr.New("ErrSourcePathNotFound")
	}
	if !op.Stat(path.Dir(recycleBinDTO.SourcePath)) {
		return buserr.New("ErrSourcePathNotFound")
	}
	if op.Stat(recycleBinDTO.SourcePath) {
		if err = op.RmRf(recycleBinDTO.SourcePath); err != nil {
			return err
		}
	}
	if err := op.Mv(filePath, recycleBinDTO.SourcePath); err != nil {
		return err
	}
	if err := cleanupTrashInfoByEntryPath(filePath); err != nil {
		global.LOG.Warnf("cleanup trashinfo failed for %s: %v", filePath, err)
	}
	return nil
}

func (r RecycleBinService) Clear() error {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return err
	}
	op := files.NewFileOp()
	for _, p := range partitions {
		dir := path.Join(p.Mountpoint, recycleBinClashDir)
		if !op.Stat(dir) {
			continue
		}
		newDir := path.Join(p.Mountpoint, "1panel_clash")
		if err := op.Mv(dir, newDir); err != nil {
			return err
		}
		go func(target string) {
			defer func() {
				if r := recover(); r != nil {
					global.LOG.Warnf("clear recycle bin panic on %s: %v", target, r)
				}
			}()
			_ = op.DeleteDir(target)
		}(newDir)
	}
	return nil
}

func getClashDir(realPath string) (string, error) {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return "", err
	}
	for _, p := range partitions {
		if p.Mountpoint == "/" {
			continue
		}
		if strings.HasPrefix(realPath, p.Mountpoint) {
			clashDir := path.Join(p.Mountpoint, recycleBinClashDir)
			if err = createClashDir(clashDir); err != nil {
				return "", err
			}
			return clashDir, nil
		}
	}
	return global.Dir.RecycleBinDir, createClashDir(global.Dir.RecycleBinDir)
}

func createClashDir(clashDir string) error {
	op := files.NewFileOp()
	if !op.Stat(clashDir) {
		if err := op.CreateDir(clashDir, constant.DirPerm); err != nil {
			return err
		}
	}
	return nil
}

func ensureTrashDirs(clashRoot string) (string, string, error) {
	filesDir := path.Join(clashRoot, recycleBinFilesSubdir)
	infoDir := path.Join(clashRoot, recycleBinInfoSubdir)
	op := files.NewFileOp()
	if !op.Stat(filesDir) {
		if err := op.CreateDir(filesDir, constant.DirPerm); err != nil {
			return "", "", err
		}
	}
	if !op.Stat(infoDir) {
		if err := op.CreateDir(infoDir, constant.DirPerm); err != nil {
			return "", "", err
		}
	}
	return filesDir, infoDir, nil
}

func trashFilesDir(clashRoot string) string {
	return path.Join(clashRoot, recycleBinFilesSubdir)
}

func trashInfoDir(clashRoot string) string {
	return path.Join(clashRoot, recycleBinInfoSubdir)
}

func trashInfoPath(infoDir, entryName string) string {
	return path.Join(infoDir, entryName+trashInfoSuffix)
}

type trashInfo struct {
	Path         string
	DeletionDate time.Time
	Size         int64
	IsDir        bool
}

func writeTrashInfo(dstPath string, info trashInfo) error {
	cfg := ini.Empty()
	section, err := cfg.NewSection(trashInfoSection)
	if err != nil {
		return err
	}
	if _, err := section.NewKey("Path", url.PathEscape(info.Path)); err != nil {
		return err
	}
	if _, err := section.NewKey("DeletionDate", info.DeletionDate.Format(trashInfoTimeLayout)); err != nil {
		return err
	}
	if _, err := section.NewKey("Size", strconv.FormatInt(info.Size, 10)); err != nil {
		return err
	}
	if _, err := section.NewKey("IsDir", strconv.FormatBool(info.IsDir)); err != nil {
		return err
	}

	var buf bytes.Buffer
	if _, err := cfg.WriteTo(&buf); err != nil {
		return err
	}
	tmp := dstPath + ".tmp"
	if err := os.WriteFile(tmp, buf.Bytes(), constant.FilePerm); err != nil {
		return err
	}
	return os.Rename(tmp, dstPath)
}

func readTrashInfo(srcPath string) (*trashInfo, error) {
	cfg, err := ini.Load(srcPath)
	if err != nil {
		return nil, err
	}
	section, err := cfg.GetSection(trashInfoSection)
	if err != nil {
		return nil, err
	}
	rawPath := section.Key("Path").Value()
	decodedPath, err := url.PathUnescape(rawPath)
	if err != nil {
		decodedPath = rawPath
	}
	info := &trashInfo{
		Path:  decodedPath,
		IsDir: section.Key("IsDir").MustBool(false),
	}
	if deletionStr := section.Key("DeletionDate").Value(); deletionStr != "" {
		if t, parseErr := time.ParseInLocation(trashInfoTimeLayout, deletionStr, time.Local); parseErr == nil {
			info.DeletionDate = t
		}
	}
	if sizeStr := section.Key("Size").Value(); sizeStr != "" {
		if sz, sizeErr := strconv.ParseInt(sizeStr, 10, 64); sizeErr == nil {
			info.Size = sz
		}
	}
	return info, nil
}

func collectTrashEntries(clashRoot string) []response.RecycleBinDTO {
	filesDir := trashFilesDir(clashRoot)
	infoDir := trashInfoDir(clashRoot)
	op := files.NewFileOp()
	if !op.Stat(filesDir) {
		return nil
	}
	entries, err := os.ReadDir(filesDir)
	if err != nil {
		global.LOG.Warnf("read recycle files dir %s failed: %v", filesDir, err)
		return nil
	}
	var result []response.RecycleBinDTO
	seen := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		seen[entry.Name()] = struct{}{}
		dto, err := buildTrashDTO(filesDir, infoDir, entry)
		if err != nil {
			global.LOG.Warnf("build recycle dto for %s failed: %v", entry.Name(), err)
			continue
		}
		result = append(result, *dto)
	}
	pruneOrphanTrashInfo(infoDir, seen)
	return result
}

func collectLegacyEntries(clashRoot string) []response.RecycleBinDTO {
	op := files.NewFileOp()
	if !op.Stat(clashRoot) {
		return nil
	}
	entries, err := os.ReadDir(clashRoot)
	if err != nil {
		return nil
	}
	var result []response.RecycleBinDTO
	for _, entry := range entries {
		name := entry.Name()
		if name == recycleBinFilesSubdir || name == recycleBinInfoSubdir {
			continue
		}
		if !strings.HasPrefix(name, "_1p_") {
			continue
		}
		dto, err := getRecycleBinDTOFromName(name)
		if err != nil {
			continue
		}
		dto.IsDir = entry.IsDir()
		dto.From = clashRoot
		result = append(result, *dto)
	}
	return result
}

func buildTrashDTO(filesDir, infoDir string, entry os.DirEntry) (*response.RecycleBinDTO, error) {
	entryName := entry.Name()
	infoPath := trashInfoPath(infoDir, entryName)
	if info, err := readTrashInfo(infoPath); err == nil {
		return trashInfoToDTO(entryName, filesDir, info), nil
	} else if !os.IsNotExist(err) {
		global.LOG.Warnf("read trashinfo %s failed: %v", infoPath, err)
	}

	entryPath := path.Join(filesDir, entryName)
	fi, err := os.Stat(entryPath)
	if err != nil {
		return nil, err
	}
	size := fi.Size()
	if fi.IsDir() {
		if sz, sizeErr := files.NewFileOp().GetDirSize(entryPath); sizeErr == nil {
			size = int64(sz)
		}
	}
	return &response.RecycleBinDTO{
		Name:       entryName,
		Size:       clampToInt(size),
		Type:       "file",
		DeleteTime: fi.ModTime(),
		RName:      entryName,
		SourcePath: "",
		IsDir:      fi.IsDir(),
		From:       filesDir,
	}, nil
}

func trashInfoToDTO(entryName, filesDir string, info *trashInfo) *response.RecycleBinDTO {
	return &response.RecycleBinDTO{
		Name:       path.Base(info.Path),
		Size:       clampToInt(info.Size),
		Type:       "file",
		DeleteTime: info.DeletionDate,
		RName:      entryName,
		SourcePath: info.Path,
		IsDir:      info.IsDir,
		From:       filesDir,
	}
}

func pruneOrphanTrashInfo(infoDir string, validEntries map[string]struct{}) {
	entries, err := os.ReadDir(infoDir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), trashInfoSuffix) {
			continue
		}
		base := strings.TrimSuffix(entry.Name(), trashInfoSuffix)
		if _, ok := validEntries[base]; ok {
			continue
		}
		orphan := path.Join(infoDir, entry.Name())
		if err := os.Remove(orphan); err != nil {
			global.LOG.Warnf("remove orphan trashinfo %s failed: %v", orphan, err)
		}
	}
}

func loadRecycleBinDTO(from, name string) (*response.RecycleBinDTO, error) {
	if path.Base(from) == recycleBinFilesSubdir {
		infoDir := trashInfoDir(path.Dir(from))
		info, err := readTrashInfo(trashInfoPath(infoDir, name))
		if err == nil {
			return trashInfoToDTO(name, from, info), nil
		}
		if !os.IsNotExist(err) {
			global.LOG.Warnf("read trashinfo for %s failed: %v", name, err)
		}
	}

	dto, err := getRecycleBinDTOFromName(name)
	if err != nil {
		return nil, err
	}
	dto.From = from
	return dto, nil
}

func cleanupTrashInfoByEntryPath(filePath string) error {
	cleaned := path.Clean(filePath)
	parent := path.Dir(cleaned)
	if path.Base(parent) != recycleBinFilesSubdir {
		return nil
	}
	clashRoot := path.Dir(parent)
	if path.Base(clashRoot) != recycleBinClashDir {
		return nil
	}
	infoPath := trashInfoPath(trashInfoDir(clashRoot), path.Base(cleaned))
	if err := os.Remove(infoPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func allocateTrashEntryName(filesDir, infoDir, sourcePath string) string {
	ext := path.Ext(path.Base(sourcePath))
	op := files.NewFileOp()
	for i := 0; i < 5; i++ {
		name := strings.ReplaceAll(common.GetUuid(), "-", "") + ext
		if op.Stat(path.Join(filesDir, name)) {
			continue
		}
		if op.Stat(trashInfoPath(infoDir, name)) {
			continue
		}
		return name
	}
	return fmt.Sprintf("%s-%d%s", strings.ReplaceAll(common.GetUuid(), "-", ""), time.Now().UnixNano(), ext)
}

func clampToInt(v int64) int {
	if v > math.MaxInt {
		return math.MaxInt
	}
	if v < math.MinInt {
		return math.MinInt
	}
	return int(v)
}

func getRecycleBinDTOFromName(filename string) (*response.RecycleBinDTO, error) {
	matches := re.GetRegex(re.RecycleBinFilePattern).FindStringSubmatch(filename)
	if len(matches) != 4 {
		return nil, fmt.Errorf("invalid filename format")
	}
	sourcePath := "/" + strings.ReplaceAll(matches[1], "_1p_", "/")
	size, err := strconv.ParseInt(matches[2], 10, 64)
	if err != nil {
		return nil, err
	}
	if size < math.MinInt || size > math.MaxInt {
		return nil, fmt.Errorf("size out of int range")
	}

	deleteTime, err := strconv.ParseInt(matches[3], 10, 64)
	if err != nil {
		return nil, err
	}
	return &response.RecycleBinDTO{
		Name:       path.Base(sourcePath),
		Size:       int(size),
		Type:       "file",
		DeleteTime: time.Unix(deleteTime, 0),
		SourcePath: sourcePath,
		RName:      filename,
	}, nil
}
