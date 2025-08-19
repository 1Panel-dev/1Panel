package service

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/build"
	"github.com/docker/docker/api/types/filters"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	fileUtils "github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/google/uuid"
)

const (
	upgradePath      = "1panel/tmp/upgrade"
	snapshotTmpPath  = "1panel/tmp/system"
	rollbackPath     = "1panel/tmp"
	cachePath        = "1panel/cache"
	oldOriginalPath  = "original"
	oldAppBackupPath = "1panel/resource/apps_bak"
	oldDownloadPath  = "1panel/tmp/download"
	oldUpgradePath   = "1panel/tmp"
	tmpUploadPath    = "1panel/tmp/upload"
	uploadPath       = "1panel/uploads"
	downloadPath     = "1panel/download"
)

func (u *DeviceService) Scan() dto.CleanData {
	var (
		SystemClean dto.CleanData
		treeData    []dto.CleanTree
	)
	fileOp := fileUtils.NewFileOp()

	originalPath := path.Join(global.Dir.BaseDir, "1panel_original")
	originalSize, _ := fileOp.GetDirSize(originalPath)
	treeData = append(treeData, dto.CleanTree{
		ID:          uuid.NewString(),
		Label:       "1panel_original",
		Size:        uint64(originalSize),
		IsCheck:     originalSize > 0,
		IsRecommend: true,
		Type:        "1panel_original",
		Children:    loadTreeWithDir(true, "1panel_original", originalPath, fileOp),
	})

	upgradePath := path.Join(global.Dir.BaseDir, upgradePath)
	upgradeSize, _ := fileOp.GetDirSize(upgradePath)
	upgradeTree := dto.CleanTree{
		ID:          uuid.NewString(),
		Label:       "upgrade",
		Size:        uint64(upgradeSize),
		IsCheck:     false,
		IsRecommend: true,
		Type:        "upgrade",
		Children:    loadTreeWithDir(true, "upgrade", upgradePath, fileOp),
	}
	if len(upgradeTree.Children) != 0 {
		sort.Slice(upgradeTree.Children, func(i, j int) bool {
			return common.CompareVersion(upgradeTree.Children[i].Label, upgradeTree.Children[j].Label)
		})
		upgradeTree.Children[0].IsCheck = false
		upgradeTree.Children[0].IsRecommend = false
	}
	treeData = append(treeData, upgradeTree)

	tmpBackupTree := loadTmpBackupTree(fileOp)
	tmpBackupSize := uint64(0)
	for _, tmp := range tmpBackupTree {
		tmpBackupSize += tmp.Size
	}
	treeData = append(treeData, dto.CleanTree{
		ID:          uuid.NewString(),
		Label:       "backup",
		Size:        tmpBackupSize,
		IsCheck:     tmpBackupSize > 0,
		IsRecommend: true,
		Type:        "backup",
		Children:    tmpBackupTree,
	})

	rollBackTree := loadRollBackTree(fileOp)
	rollbackSize := uint64(0)
	for _, rollback := range rollBackTree {
		rollbackSize += rollback.Size
	}
	treeData = append(treeData, dto.CleanTree{
		ID:          uuid.NewString(),
		Label:       "rollback",
		Size:        rollbackSize,
		IsCheck:     rollbackSize > 0,
		IsRecommend: true,
		Type:        "rollback",
		Children:    rollBackTree,
	})

	cachePath := path.Join(global.Dir.BaseDir, cachePath)
	cacheSize, _ := fileOp.GetDirSize(cachePath)
	treeData = append(treeData, dto.CleanTree{
		ID:          uuid.NewString(),
		Label:       "cache",
		Size:        uint64(cacheSize),
		IsCheck:     false,
		IsRecommend: false,
		Type:        "cache",
	})
	SystemClean.SystemClean = treeData

	uploadTreeData := loadUploadTree(fileOp)
	SystemClean.UploadClean = append(SystemClean.UploadClean, uploadTreeData...)

	downloadTreeData := loadDownloadTree(fileOp)
	SystemClean.DownloadClean = append(SystemClean.DownloadClean, downloadTreeData...)

	logTree := loadLogTree(fileOp)
	SystemClean.SystemLogClean = append(SystemClean.SystemLogClean, logTree...)

	containerTree := loadContainerTree()
	SystemClean.ContainerClean = append(SystemClean.ContainerClean, containerTree...)

	return SystemClean
}

func (u *DeviceService) Clean(req []dto.Clean) {
	size := uint64(0)
	restart := false
	for _, item := range req {
		size += item.Size
		switch item.TreeType {
		case "1panel_original":
			dropFileOrDir(path.Join(global.Dir.BaseDir, "1panel_original", item.Name))

		case "upgrade":
			dropFileOrDir(path.Join(global.Dir.BaseDir, upgradePath, item.Name))

		case "backup":
			dropFileOrDir(path.Join(global.Dir.LocalBackupDir, "tmp/app"))
			dropFileOrDir(path.Join(global.Dir.LocalBackupDir, "tmp/database"))
			dropFileOrDir(path.Join(global.Dir.LocalBackupDir, "tmp/website"))
			dropFileOrDir(path.Join(global.Dir.LocalBackupDir, "tmp/directory"))
			dropFileOrDir(path.Join(global.Dir.LocalBackupDir, "tmp/log"))
			dropFileOrDir(path.Join(global.Dir.LocalBackupDir, "tmp/system"))

		case "rollback":
			dropFileOrDir(path.Join(global.Dir.BaseDir, rollbackPath, "app"))
			dropFileOrDir(path.Join(global.Dir.BaseDir, rollbackPath, "database"))
			dropFileOrDir(path.Join(global.Dir.BaseDir, rollbackPath, "website"))
		case "rollback_app":
			dropFileOrDir(path.Join(global.Dir.BaseDir, rollbackPath, "app", item.Name))
		case "rollback_database":
			dropFileOrDir(path.Join(global.Dir.BaseDir, rollbackPath, "database", item.Name))
		case "rollback_website":
			dropFileOrDir(path.Join(global.Dir.BaseDir, rollbackPath, "website", item.Name))

		case "cache":
			dropFileOrDir(path.Join(global.Dir.BaseDir, cachePath, item.Name))
			restart = true

		case "upload":
			dropFileOrDir(path.Join(global.Dir.BaseDir, uploadPath, item.Name))
			if len(item.Name) == 0 {
				dropFileOrDir(path.Join(global.Dir.BaseDir, tmpUploadPath))
			}
		case "upload_tmp":
			dropFileOrDir(path.Join(global.Dir.BaseDir, tmpUploadPath, item.Name))
		case "upload_app":
			dropFileOrDir(path.Join(global.Dir.BaseDir, uploadPath, "app", item.Name))
		case "upload_database":
			dropFileOrDir(path.Join(global.Dir.BaseDir, uploadPath, "database", item.Name))
		case "upload_website":
			dropFileOrDir(path.Join(global.Dir.BaseDir, uploadPath, "website", item.Name))
		case "upload_directory":
			dropFileOrDir(path.Join(global.Dir.BaseDir, uploadPath, "directory", item.Name))
		case "download":
			dropFileOrDir(path.Join(global.Dir.BaseDir, downloadPath, item.Name))
		case "download_app":
			dropFileOrDir(path.Join(global.Dir.BaseDir, downloadPath, "app", item.Name))
		case "download_database":
			dropFileOrDir(path.Join(global.Dir.BaseDir, downloadPath, "database", item.Name))
		case "download_website":
			dropFileOrDir(path.Join(global.Dir.BaseDir, downloadPath, "website", item.Name))
		case "download_directory":
			dropFileOrDir(path.Join(global.Dir.BaseDir, downloadPath, "directory", item.Name))

		case "system_log":
			if len(item.Name) == 0 {
				files, _ := os.ReadDir(global.Dir.LogDir)
				if len(files) == 0 {
					continue
				}
				for _, file := range files {
					if file.Name() == "1Panel-Core.log" || file.Name() == "1Panel.log" || file.IsDir() {
						continue
					}
					dropFileOrDir(path.Join(global.Dir.LogDir, file.Name()))
				}
			} else {
				dropFileOrDir(path.Join(global.Dir.LogDir, item.Name))
			}
		case "task_log":
			if len(item.Name) == 0 {
				files, _ := os.ReadDir(path.Join(global.Dir.TaskDir))
				if len(files) == 0 {
					continue
				}
				for _, file := range files {
					if file.Name() == "ssl" || !file.IsDir() {
						continue
					}
					dropFileOrDir(path.Join(global.Dir.TaskDir, file.Name()))
				}
				_ = taskRepo.DeleteAll()
			} else {
				pathItem := path.Join(global.Dir.TaskDir, item.Name)
				dropFileOrDir(pathItem)
				if len(item.Name) != 0 {
					_ = taskRepo.Delete(repo.WithByType(item.Name))
				}
			}
		case "script":
			dropFileOrDir(path.Join(global.Dir.TmpDir, "script", item.Name))
		case "images":
			dropImages()
		case "containers":
			dropContainers()
		case "volumes":
			dropVolumes()
		case "build_cache":
			dropBuildCache()
		}
	}

	_ = settingRepo.Update("LastCleanTime", time.Now().Format(constant.DateTimeLayout))
	_ = settingRepo.Update("LastCleanSize", fmt.Sprintf("%v", size))
	_ = settingRepo.Update("LastCleanData", fmt.Sprintf("%v", len(req)))

	if restart {
		go common.RestartService(false, true, false)
	}
}

func doSystemClean(taskItem *task.Task) func(t *task.Task) error {
	return func(t *task.Task) error {
		size := int64(0)
		fileCount := 0
		dropWithTask(path.Join(global.Dir.BaseDir, "1panel_original"), taskItem, &size, &fileCount)

		upgradePath := path.Join(global.Dir.BaseDir, upgradePath)
		upgradeFiles, _ := os.ReadDir(upgradePath)
		if len(upgradeFiles) != 0 {
			sort.Slice(upgradeFiles, func(i, j int) bool {
				return upgradeFiles[i].Name() > upgradeFiles[j].Name()
			})
			for i := 0; i < len(upgradeFiles); i++ {
				if i != 0 {
					dropWithTask(path.Join(upgradePath, upgradeFiles[i].Name()), taskItem, &size, &fileCount)
				}
			}
		}

		dropWithTask(path.Join(global.Dir.BaseDir, snapshotTmpPath), taskItem, &size, &fileCount)
		dropWithTask(path.Join(global.Dir.LocalBackupDir, "system"), taskItem, &size, &fileCount)

		dropWithTask(path.Join(global.Dir.BaseDir, rollbackPath, "app"), taskItem, &size, &fileCount)
		dropWithTask(path.Join(global.Dir.BaseDir, rollbackPath, "website"), taskItem, &size, &fileCount)
		dropWithTask(path.Join(global.Dir.BaseDir, rollbackPath, "database"), taskItem, &size, &fileCount)

		dropWithTask(path.Join(global.Dir.BaseDir, oldOriginalPath), taskItem, &size, &fileCount)
		dropWithTask(path.Join(global.Dir.BaseDir, oldAppBackupPath), taskItem, &size, &fileCount)
		dropWithTask(path.Join(global.Dir.BaseDir, oldDownloadPath), taskItem, &size, &fileCount)
		oldUpgradePath := path.Join(global.Dir.BaseDir, oldUpgradePath)
		oldUpgradeFiles, _ := os.ReadDir(oldUpgradePath)
		if len(oldUpgradeFiles) != 0 {
			for i := 0; i < len(oldUpgradeFiles); i++ {
				dropWithTask(path.Join(oldUpgradePath, oldUpgradeFiles[i].Name()), taskItem, &size, &fileCount)
			}
		}

		dropWithTask(path.Join(global.Dir.BaseDir, tmpUploadPath), taskItem, &size, &fileCount)
		dropWithExclude(path.Join(global.Dir.BaseDir, uploadPath), []string{"theme"}, taskItem, &size, &fileCount)
		dropWithTask(path.Join(global.Dir.BaseDir, downloadPath), taskItem, &size, &fileCount)

		logFiles, _ := os.ReadDir(global.Dir.LogDir)
		if len(logFiles) != 0 {
			for i := 0; i < len(logFiles); i++ {
				if logFiles[i].IsDir() {
					continue
				}
				if logFiles[i].Name() != "1Panel.log" && logFiles[i].Name() != "1Panel-Core.log" {
					dropWithTask(path.Join(global.Dir.LogDir, logFiles[i].Name()), taskItem, &size, &fileCount)
				}
			}
		}
		timeNow := time.Now().Format(constant.DateTimeLayout)
		if fileCount != 0 {
			taskItem.Log(i18n.GetMsgWithMap("FileDropSum", map[string]interface{}{"size": common.LoadSizeUnit2F(float64(size)), "count": fileCount}))
		}

		_ = settingRepo.Update("LastCleanTime", timeNow)
		_ = settingRepo.Update("LastCleanSize", fmt.Sprintf("%v", size))
		_ = settingRepo.Update("LastCleanData", fmt.Sprintf("%v", fileCount))

		return nil
	}
}

func loadTmpBackupTree(fileOp fileUtils.FileOp) []dto.CleanTree {
	var treeData []dto.CleanTree
	treeData = loadTreeWithCheck(treeData, path.Join(global.Dir.LocalBackupDir, "tmp/app"), "tmp_backup_app", fileOp)
	treeData = loadTreeWithCheck(treeData, path.Join(global.Dir.LocalBackupDir, "tmp/website"), "tmp_backup_website", fileOp)
	treeData = loadTreeWithCheck(treeData, path.Join(global.Dir.LocalBackupDir, "tmp/database"), "tmp_backup_database", fileOp)
	treeData = loadTreeWithCheck(treeData, path.Join(global.Dir.LocalBackupDir, "tmp/system"), "tmp_backup_snapshot", fileOp)
	treeData = loadTreeWithCheck(treeData, path.Join(global.Dir.LocalBackupDir, "tmp/directory"), "tmp_backup_directory", fileOp)
	treeData = loadTreeWithCheck(treeData, path.Join(global.Dir.LocalBackupDir, "tmp/log"), "tmp_backup_log", fileOp)
	return treeData
}

func loadRollBackTree(fileOp fileUtils.FileOp) []dto.CleanTree {
	var treeData []dto.CleanTree
	treeData = loadTreeWithCheck(treeData, path.Join(global.Dir.BaseDir, rollbackPath, "app"), "rollback_app", fileOp)
	treeData = loadTreeWithCheck(treeData, path.Join(global.Dir.BaseDir, rollbackPath, "website"), "rollback_website", fileOp)
	treeData = loadTreeWithCheck(treeData, path.Join(global.Dir.BaseDir, rollbackPath, "database"), "rollback_database", fileOp)

	return treeData
}

func loadUploadTree(fileOp fileUtils.FileOp) []dto.CleanTree {
	var treeData []dto.CleanTree
	treeData = loadTreeWithCheck(treeData, path.Join(global.Dir.BaseDir, tmpUploadPath), "upload_tmp", fileOp)
	treeData = loadTreeWithCheck(treeData, path.Join(global.Dir.BaseDir, uploadPath, "app"), "upload_app", fileOp)
	treeData = loadTreeWithCheck(treeData, path.Join(global.Dir.BaseDir, uploadPath, "website"), "upload_website", fileOp)
	treeData = loadTreeWithCheck(treeData, path.Join(global.Dir.BaseDir, uploadPath, "database"), "upload_database", fileOp)

	path5 := path.Join(global.Dir.BaseDir, uploadPath)
	uploadTreeData := loadTreeWithAllFile(true, path5, "upload", path5, fileOp)
	treeData = append(treeData, uploadTreeData...)

	return treeData
}

func loadDownloadTree(fileOp fileUtils.FileOp) []dto.CleanTree {
	var treeData []dto.CleanTree
	treeData = loadTreeWithCheck(treeData, path.Join(global.Dir.BaseDir, downloadPath, "app"), "download_app", fileOp)
	treeData = loadTreeWithCheck(treeData, path.Join(global.Dir.BaseDir, downloadPath, "website"), "download_website", fileOp)
	treeData = loadTreeWithCheck(treeData, path.Join(global.Dir.BaseDir, downloadPath, "database"), "download_database", fileOp)

	path5 := path.Join(global.Dir.BaseDir, downloadPath)
	uploadTreeData := loadTreeWithAllFile(true, path5, "download", path5, fileOp)
	treeData = append(treeData, uploadTreeData...)

	return treeData
}

func loadLogTree(fileOp fileUtils.FileOp) []dto.CleanTree {
	var treeData []dto.CleanTree
	path1 := path.Join(global.Dir.LogDir)
	list1 := loadTreeWithAllFile(true, path1, "system_log", path1, fileOp)
	size := uint64(0)
	for _, file := range list1 {
		size += file.Size
	}
	treeData = append(treeData, dto.CleanTree{ID: uuid.NewString(), Label: "system_log", Size: uint64(size), Children: list1, Type: "system_log", IsRecommend: true})

	path2 := path.Join(global.Dir.TaskDir)
	list2 := loadTreeWithDir(false, "task_log", path2, fileOp)
	size2, _ := fileOp.GetDirSize(path2)
	treeData = append(treeData, dto.CleanTree{ID: uuid.NewString(), Label: "task_log", Size: uint64(size2), Children: list2, Type: "task_log"})

	path3 := path.Join(global.Dir.TmpDir, "script")
	list3 := loadTreeWithAllFile(true, path3, "script", path3, fileOp)
	size3, _ := fileOp.GetDirSize(path3)
	treeData = append(treeData, dto.CleanTree{ID: uuid.NewString(), Label: "script", Size: uint64(size3), Children: list3, Type: "script", IsRecommend: true})
	return treeData
}

func loadContainerTree() []dto.CleanTree {
	var treeData []dto.CleanTree
	client, err := docker.NewDockerClient()
	if err != nil {
		return treeData
	}
	diskUsage, err := client.DiskUsage(context.Background(), types.DiskUsageOptions{})
	if err != nil {
		return treeData
	}
	imageSize := uint64(0)
	for _, file := range diskUsage.Images {
		if file.Containers == 0 {
			imageSize += uint64(file.Size)
		}
	}
	treeData = append(treeData, dto.CleanTree{ID: uuid.NewString(), Label: "container_images", Size: imageSize, Children: nil, Type: "images", IsRecommend: true})

	containerSize := uint64(0)
	for _, file := range diskUsage.Containers {
		if file.State != "running" {
			containerSize += uint64(file.SizeRw)
		}
	}
	treeData = append(treeData, dto.CleanTree{ID: uuid.NewString(), Label: "container_containers", Size: containerSize, Children: nil, Type: "containers", IsRecommend: true})

	volumeSize := uint64(0)
	for _, file := range diskUsage.Volumes {
		if file.UsageData.RefCount <= 0 {
			volumeSize += uint64(file.UsageData.Size)
		}
	}
	treeData = append(treeData, dto.CleanTree{ID: uuid.NewString(), Label: "container_volumes", Size: volumeSize, IsCheck: volumeSize > 0, Children: nil, Type: "volumes", IsRecommend: true})

	var buildCacheTotalSize int64
	for _, cache := range diskUsage.BuildCache {
		if cache.Type == "source.local" {
			buildCacheTotalSize += cache.Size
		}
	}
	treeData = append(treeData, dto.CleanTree{ID: uuid.NewString(), Label: "build_cache", Size: uint64(buildCacheTotalSize), IsCheck: buildCacheTotalSize > 0, Type: "build_cache", IsRecommend: true})
	return treeData
}

func loadTreeWithCheck(treeData []dto.CleanTree, pathItem, treeType string, fileOp fileUtils.FileOp) []dto.CleanTree {
	size, _ := fileOp.GetDirSize(pathItem)
	if size == 0 {
		return treeData
	}
	list := loadTreeWithAllFile(true, pathItem, treeType, pathItem, fileOp)
	treeData = append(treeData, dto.CleanTree{ID: uuid.NewString(), Label: treeType, Size: uint64(size), IsCheck: size > 0, Children: list, Type: treeType, IsRecommend: true})
	return treeData
}

func loadTreeWithDir(isCheck bool, treeType, pathItem string, fileOp fileUtils.FileOp) []dto.CleanTree {
	var lists []dto.CleanTree
	files, err := os.ReadDir(pathItem)
	if err != nil {
		return lists
	}
	for _, file := range files {
		if file.Name() == "ssl" {
			continue
		}
		if file.IsDir() {
			size, err := fileOp.GetDirSize(path.Join(pathItem, file.Name()))
			if err != nil {
				continue
			}
			item := dto.CleanTree{
				ID:          uuid.NewString(),
				Label:       file.Name(),
				Type:        treeType,
				Size:        uint64(size),
				Name:        strings.TrimPrefix(file.Name(), "/"),
				IsCheck:     isCheck,
				IsRecommend: isCheck,
			}
			lists = append(lists, item)
		}
	}
	return lists
}

func loadTreeWithAllFile(isCheck bool, originalPath, treeType, pathItem string, fileOp fileUtils.FileOp) []dto.CleanTree {
	var lists []dto.CleanTree

	files, err := os.ReadDir(pathItem)
	if err != nil {
		return lists
	}
	for _, file := range files {
		if treeType == "upload" && (file.Name() == "theme" && file.IsDir()) {
			continue
		}
		if treeType == "system_log" && (file.Name() == "1Panel-Core.log" || file.Name() == "1Panel.log" || file.IsDir()) {
			continue
		}
		if (treeType == "upload" || treeType == "download") && file.IsDir() && (file.Name() == "app" || file.Name() == "database" || file.Name() == "website" || file.Name() == "directory") {
			continue
		}
		size := uint64(0)
		name := strings.TrimPrefix(path.Join(pathItem, file.Name()), originalPath+"/")
		if file.IsDir() {
			sizeItem, err := fileOp.GetDirSize(path.Join(pathItem, file.Name()))
			if err != nil {
				continue
			}
			size = uint64(sizeItem)
		} else {
			fileInfo, err := file.Info()
			if err != nil {
				continue
			}
			size = uint64(fileInfo.Size())
		}
		item := dto.CleanTree{
			ID:          uuid.NewString(),
			Label:       file.Name(),
			Type:        treeType,
			Size:        uint64(size),
			Name:        name,
			IsCheck:     isCheck,
			IsRecommend: isCheck,
		}
		if file.IsDir() {
			item.Children = loadTreeWithAllFile(isCheck, originalPath, treeType, path.Join(pathItem, file.Name()), fileOp)
		}
		lists = append(lists, item)
	}
	return lists
}

func dropFileOrDir(itemPath string) {
	global.LOG.Debugf("drop file %s", itemPath)
	if err := os.RemoveAll(itemPath); err != nil {
		global.LOG.Errorf("drop file %s failed, err %v", itemPath, err)
	}
}

func dropBuildCache() {
	client, err := docker.NewDockerClient()
	if err != nil {
		global.LOG.Errorf("do not get docker client")
	}
	opts := build.CachePruneOptions{}
	opts.All = true
	_, err = client.BuildCachePrune(context.Background(), opts)
	if err != nil {
		global.LOG.Errorf("drop build cache failed, err %v", err)
	}
}

func dropImages() {
	client, err := docker.NewDockerClient()
	if err != nil {
		global.LOG.Errorf("do not get docker client")
	}
	pruneFilters := filters.NewArgs()
	pruneFilters.Add("dangling", "false")
	_, err = client.ImagesPrune(context.Background(), pruneFilters)
	if err != nil {
		global.LOG.Errorf("drop images failed, err %v", err)
	}
}

func dropContainers() {
	client, err := docker.NewDockerClient()
	if err != nil {
		global.LOG.Errorf("do not get docker client")
	}
	pruneFilters := filters.NewArgs()
	_, err = client.ContainersPrune(context.Background(), pruneFilters)
	if err != nil {
		global.LOG.Errorf("drop containers failed, err %v", err)
	}
}

func dropVolumes() {
	client, err := docker.NewDockerClient()
	if err != nil {
		global.LOG.Errorf("do not get docker client")
	}
	pruneFilters := filters.NewArgs()
	versions, err := client.ServerVersion(context.Background())
	if err != nil {
		global.LOG.Errorf("do not get docker api versions")
	}
	if common.ComparePanelVersion(versions.APIVersion, "1.42") {
		pruneFilters.Add("all", "true")
	}
	_, err = client.VolumesPrune(context.Background(), pruneFilters)
	if err != nil {
		global.LOG.Errorf("drop volumes failed, err %v", err)
	}
}

func dropWithExclude(pathToDelete string, excludeSubDirs []string, taskItem *task.Task, size *int64, count *int) {
	entries, err := os.ReadDir(pathToDelete)
	if err != nil {
		return
	}

	for _, entry := range entries {
		name := entry.Name()
		fullPath := filepath.Join(pathToDelete, name)
		excluded := false
		for _, ex := range excludeSubDirs {
			if name == ex {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}
		dropWithTask(fullPath, taskItem, size, count)
	}
}

func dropWithTask(itemPath string, taskItem *task.Task, size *int64, count *int) {
	itemSize := int64(0)
	itemCount := 0
	scanFile(itemPath, &itemSize, &itemCount)
	*size += itemSize
	*count += itemCount
	if err := os.RemoveAll(itemPath); err != nil {
		taskItem.Log(i18n.GetMsgWithDetail("FileDropFailed", err.Error()))
		return
	}
	if itemCount != 0 {
		taskItem.Log(i18n.GetMsgWithMap("FileDropSuccess", map[string]interface{}{"count": itemCount, "size": common.LoadSizeUnit2F(float64(itemSize))}))
	}
}

func scanFile(pathItem string, size *int64, count *int) {
	files, _ := os.ReadDir(pathItem)
	for _, f := range files {
		if f.IsDir() {
			scanFile(path.Join(pathItem, f.Name()), size, count)
		} else {
			fileInfo, err := f.Info()
			if err != nil {
				continue
			}
			*count++
			*size += fileInfo.Size()
		}
	}
}
