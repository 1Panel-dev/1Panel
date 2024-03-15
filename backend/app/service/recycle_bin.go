package service

import (
	"fmt"
	"math"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/app/dto/response"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"github.com/shirou/gopsutil/v3/disk"
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
			if strings.HasPrefix(file.Name(), "_1p_") {
				recycleDTO, err := getRecycleBinDTOFromName(file.Name())
				recycleDTO.IsDir = file.IsDir()
				recycleDTO.From = dir
				if err == nil {
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
	op := files.NewFileOp()
	if !op.Stat(create.SourcePath) {
		return buserr.New(constant.ErrLinkPathNotFound)
	}
	clashDir, err := getClashDir(create.SourcePath)
	if err != nil {
		return err
	}
	paths := strings.Split(create.SourcePath, "/")
	rNamePre := strings.Join(paths, "_1p_")
	deleteTime := time.Now()
	openFile, err := op.OpenFile(create.SourcePath)
	if err != nil {
		return err
	}
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

	rName := fmt.Sprintf("_1p_%s%s_p_%d_%d", "file", rNamePre, size, deleteTime.Unix())
	return op.Mv(create.SourcePath, path.Join(clashDir, rName))
}

func (r RecycleBinService) Reduce(reduce request.RecycleBinReduce) error {
	filePath := path.Join(reduce.From, reduce.RName)
	op := files.NewFileOp()
	if !op.Stat(filePath) {
		return buserr.New(constant.ErrLinkPathNotFound)
	}
	recycleBinDTO, err := getRecycleBinDTOFromName(reduce.RName)
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
	return op.Mv(filePath, recycleBinDTO.SourcePath)
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
	return constant.RecycleBinDir, createClashDir(constant.RecycleBinDir)
}

func createClashDir(clashDir string) error {
	op := files.NewFileOp()
	if !op.Stat(clashDir) {
		if err := op.CreateDir(clashDir, 0755); err != nil {
			return err
		}
	}
	return nil
}

func getRecycleBinDTOFromName(filename string) (*response.RecycleBinDTO, error) {
	r := regexp.MustCompile(`_1p_file_1p_(.+)_p_(\d+)_(\d+)`)
	matches := r.FindStringSubmatch(filename)
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
