package service

import (
	"encoding/json"
	"fmt"
	"math"
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
)

const recycleBinMetaSuffix = ".1panel-meta.json"

type recycleBinMeta struct {
	SourcePath string `json:"sourcePath"`
	Size       int    `json:"size"`
	DeleteTime int64  `json:"deleteTime"`
	IsDir      bool   `json:"isDir"`
}

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
	var (
		result []response.RecycleBinDTO
	)
	partitions, err := disk.Partitions(false)
	if err != nil {
		return 0, nil, err
	}
	op := files.NewFileOp()
	for _, p := range partitions {
		dir := path.Join(p.Mountpoint, ".1panel_clash")
		if !op.Stat(dir) {
			continue
		}
		clashFiles, err := os.ReadDir(dir)
		if err != nil {
			return 0, nil, err
		}
		for _, file := range clashFiles {
			if strings.HasSuffix(file.Name(), recycleBinMetaSuffix) {
				entryName := strings.TrimSuffix(file.Name(), recycleBinMetaSuffix)
				if !op.Stat(path.Join(dir, entryName)) {
					continue
				}
				recycleDTO, err := loadRecycleBinDTO(dir, entryName)
				if err == nil {
					result = append(result, *recycleDTO)
				}
				continue
			}

			if strings.HasPrefix(file.Name(), "_1p_") {
				recycleDTO, err := getRecycleBinDTOFromName(file.Name())
				if err == nil {
					recycleDTO.IsDir = file.IsDir()
					recycleDTO.From = dir
					result = append(result, *recycleDTO)
				}
			}
		}
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
	clashDir, err := getClashDir(create.SourcePath)
	if err != nil {
		return err
	}
	deleteTime := time.Now()
	openFile, err := op.OpenFile(create.SourcePath)
	if err != nil {
		return err
	}
	defer openFile.Close()
	fileInfo, err := openFile.Stat()
	if err != nil {
		return err
	}
	size := 0
	if fileInfo.IsDir() {
		sizeF, err := op.GetDirSize(create.SourcePath)
		if err != nil {
			return err
		}
		size = int(sizeF)
	} else {
		size = int(fileInfo.Size())
	}

	rName := buildRecycleBinEntryName(deleteTime)
	meta := recycleBinMeta{
		SourcePath: create.SourcePath,
		Size:       size,
		DeleteTime: deleteTime.Unix(),
		IsDir:      fileInfo.IsDir(),
	}
	targetPath := path.Join(clashDir, rName)
	if err := op.Mv(create.SourcePath, targetPath); err != nil {
		return err
	}
	if err := saveRecycleBinMeta(clashDir, rName, meta); err != nil {
		if rollbackErr := op.Mv(targetPath, create.SourcePath); rollbackErr != nil {
			global.LOG.Warnf("rollback recycle bin create failed for %s: %v", create.SourcePath, rollbackErr)
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
	if err := cleanupRecycleBinMetaByEntryPath(filePath); err != nil {
		global.LOG.Warnf("cleanup recycle bin metadata failed for %s: %v", filePath, err)
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
		dir := path.Join(p.Mountpoint, ".1panel_clash")
		if !op.Stat(dir) {
			continue
		}
		newDir := path.Join(p.Mountpoint, "1panel_clash")
		if err := op.Mv(dir, newDir); err != nil {
			return err
		}
		go func() {
			_ = op.DeleteDir(newDir)
		}()
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
			clashDir := path.Join(p.Mountpoint, ".1panel_clash")
			if err = createClashDir(path.Join(p.Mountpoint, ".1panel_clash")); err != nil {
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

func buildRecycleBinEntryName(deleteTime time.Time) string {
	return fmt.Sprintf("_1p_file_%d_%s", deleteTime.Unix(), strings.ReplaceAll(common.GetUuid(), "-", ""))
}

func getRecycleBinMetaPath(dir, name string) string {
	return path.Join(dir, name+recycleBinMetaSuffix)
}

func saveRecycleBinMeta(dir, name string, meta recycleBinMeta) error {
	content, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	return files.NewFileOp().SaveFile(getRecycleBinMetaPath(dir, name), string(content), constant.FilePerm)
}

func loadRecycleBinMeta(dir, name string) (*recycleBinMeta, error) {
	content, err := os.ReadFile(getRecycleBinMetaPath(dir, name))
	if err != nil {
		return nil, err
	}
	var meta recycleBinMeta
	if err := json.Unmarshal(content, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

func getRecycleBinDTOFromMeta(name, from string, meta *recycleBinMeta) *response.RecycleBinDTO {
	return &response.RecycleBinDTO{
		Name:       path.Base(meta.SourcePath),
		Size:       meta.Size,
		Type:       "file",
		DeleteTime: time.Unix(meta.DeleteTime, 0),
		SourcePath: meta.SourcePath,
		RName:      name,
		IsDir:      meta.IsDir,
		From:       from,
	}
}

func loadRecycleBinDTO(from, name string) (*response.RecycleBinDTO, error) {
	meta, err := loadRecycleBinMeta(from, name)
	if err == nil {
		return getRecycleBinDTOFromMeta(name, from, meta), nil
	}
	if !os.IsNotExist(err) {
		return nil, err
	}

	recycleDTO, err := getRecycleBinDTOFromName(name)
	if err != nil {
		return nil, err
	}
	recycleDTO.From = from
	return recycleDTO, nil
}

func cleanupRecycleBinMetaByEntryPath(filePath string) error {
	if !isRecycleBinEntryPath(filePath) {
		return nil
	}
	metaPath := getRecycleBinMetaPath(path.Dir(filePath), path.Base(filePath))
	if err := os.Remove(metaPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func isRecycleBinEntryPath(filePath string) bool {
	cleaned := path.Clean(filePath)
	return path.Base(path.Dir(cleaned)) == ".1panel_clash" && !strings.HasSuffix(cleaned, recycleBinMetaSuffix)
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
