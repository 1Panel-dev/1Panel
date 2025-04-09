package service

import (
	"context"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/docker/docker/api/types"
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
	logPath          = "1panel/log"
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
		IsCheck:     true,
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

	snapTree := loadSnapshotTree(fileOp)
	snapSize := uint64(0)
	for _, snap := range snapTree {
		snapSize += snap.Size
	}
	treeData = append(treeData, dto.CleanTree{
		ID:          uuid.NewString(),
		Label:       "snapshot",
		Size:        snapSize,
		IsCheck:     true,
		IsRecommend: true,
		Type:        "snapshot",
		Children:    snapTree,
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
		IsCheck:     true,
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

	unusedTree := loadUnusedFile(fileOp)
	unusedSize := uint64(0)
	for _, unused := range unusedTree {
		unusedSize += unused.Size
	}
	treeData = append(treeData, dto.CleanTree{
		ID:          uuid.NewString(),
		Label:       "unused",
		Size:        unusedSize,
		IsCheck:     true,
		IsRecommend: true,
		Type:        "unused",
		Children:    unusedTree,
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

		case "snapshot":
			dropFileOrDir(path.Join(global.Dir.BaseDir, snapshotTmpPath, item.Name))
			dropFileOrDir(path.Join(global.Dir.LocalBackupDir, "system", item.Name))
		case "snapshot_tmp":
			dropFileOrDir(path.Join(global.Dir.BaseDir, snapshotTmpPath, item.Name))
		case "snapshot_local":
			dropFileOrDir(path.Join(global.Dir.LocalBackupDir, "system", item.Name))

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

		case "unused":
			dropFileOrDir(path.Join(global.Dir.BaseDir, oldOriginalPath))
			dropFileOrDir(path.Join(global.Dir.BaseDir, oldAppBackupPath))
			dropFileOrDir(path.Join(global.Dir.BaseDir, oldDownloadPath))
			files, _ := os.ReadDir(path.Join(global.Dir.BaseDir, oldUpgradePath))
			if len(files) == 0 {
				continue
			}
			for _, file := range files {
				if strings.HasPrefix(file.Name(), "upgrade_") {
					dropFileOrDir(path.Join(global.Dir.BaseDir, oldUpgradePath, file.Name()))
				}
			}
		case "old_original":
			dropFileOrDir(path.Join(global.Dir.BaseDir, oldOriginalPath, item.Name))
		case "old_apps_bak":
			dropFileOrDir(path.Join(global.Dir.BaseDir, oldAppBackupPath, item.Name))
		case "old_download":
			dropFileOrDir(path.Join(global.Dir.BaseDir, oldDownloadPath, item.Name))
		case "old_upgrade":
			if len(item.Name) == 0 {
				files, _ := os.ReadDir(path.Join(global.Dir.BaseDir, oldUpgradePath))
				if len(files) == 0 {
					continue
				}
				for _, file := range files {
					if strings.HasPrefix(file.Name(), "upgrade_") {
						dropFileOrDir(path.Join(global.Dir.BaseDir, oldUpgradePath, file.Name()))
					}
				}
			} else {
				dropFileOrDir(path.Join(global.Dir.BaseDir, oldUpgradePath, item.Name))
			}

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
				files, _ := os.ReadDir(path.Join(global.Dir.BaseDir, logPath))
				if len(files) == 0 {
					continue
				}
				for _, file := range files {
					if file.Name() == "1Panel-Core.log" || file.Name() == "1Panel.log" || file.IsDir() {
						continue
					}
					dropFileOrDir(path.Join(global.Dir.BaseDir, logPath, file.Name()))
				}
			} else {
				dropFileOrDir(path.Join(global.Dir.BaseDir, logPath, item.Name))
			}
		case "task_log":
			if len(item.Name) == 0 {
				files, _ := os.ReadDir(path.Join(global.Dir.BaseDir, logPath))
				if len(files) == 0 {
					continue
				}
				for _, file := range files {
					if file.Name() == "ssl" || !file.IsDir() {
						continue
					}
					dropFileOrDir(path.Join(global.Dir.BaseDir, logPath, file.Name()))
				}
				_ = taskRepo.DeleteAll()
			} else {
				pathItem := path.Join(global.Dir.BaseDir, logPath, item.Name)
				dropFileOrDir(pathItem)
				if len(item.Name) != 0 {
					_ = taskRepo.Delete(repo.WithByType(item.Name))
				}
			}
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

func systemClean(taskItem *task.Task) error {
	taskItem.AddSubTask(i18n.GetMsgByKey("HandleSystemClean"), func(t *task.Task) error {
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
		dropWithTask(path.Join(global.Dir.BaseDir, uploadPath), taskItem, &size, &fileCount)
		dropWithTask(path.Join(global.Dir.BaseDir, downloadPath), taskItem, &size, &fileCount)

		logPath := path.Join(global.Dir.BaseDir, logPath)
		logFiles, _ := os.ReadDir(logPath)
		if len(logFiles) != 0 {
			for i := 0; i < len(logFiles); i++ {
				if logFiles[i].IsDir() {
					continue
				}
				if logFiles[i].Name() != "1Panel.log" && logFiles[i].Name() != "1Panel-Core.log" {
					dropWithTask(path.Join(logPath, logFiles[i].Name()), taskItem, &size, &fileCount)
				}
			}
		}
		timeNow := time.Now().Format(constant.DateTimeLayout)
		if fileCount != 0 {
			taskItem.LogSuccessF("%s: total clean: %s, total count: %d", timeNow, common.LoadSizeUnit2F(float64(size)), fileCount)
		}

		_ = settingRepo.Update("LastCleanTime", timeNow)
		_ = settingRepo.Update("LastCleanSize", fmt.Sprintf("%v", size))
		_ = settingRepo.Update("LastCleanData", fmt.Sprintf("%v", fileCount))

		return nil
	}, nil)
	return taskItem.Execute()
}

func loadSnapshotTree(fileOp fileUtils.FileOp) []dto.CleanTree {
	var treeData []dto.CleanTree
	path1 := path.Join(global.Dir.BaseDir, snapshotTmpPath)
	list1 := loadTreeWithAllFile(true, path1, "snapshot_tmp", path1, fileOp)
	if len(list1) != 0 {
		size, _ := fileOp.GetDirSize(path1)
		treeData = append(treeData, dto.CleanTree{ID: uuid.NewString(), Label: "snapshot_tmp", Size: uint64(size), Children: list1, Type: "snapshot_tmp", IsRecommend: true})
	}

	path2 := path.Join(global.Dir.LocalBackupDir, "system")
	list2 := loadTreeWithAllFile(true, path2, "snapshot_local", path2, fileOp)
	if len(list2) != 0 {
		size, _ := fileOp.GetDirSize(path2)
		treeData = append(treeData, dto.CleanTree{ID: uuid.NewString(), Label: "snapshot_local", Size: uint64(size), Children: list2, Type: "snapshot_local", IsRecommend: true})
	}
	return treeData
}

func loadRollBackTree(fileOp fileUtils.FileOp) []dto.CleanTree {
	var treeData []dto.CleanTree
	path1 := path.Join(global.Dir.BaseDir, rollbackPath, "app")
	list1 := loadTreeWithAllFile(true, path1, "rollback_app", path1, fileOp)
	size1, _ := fileOp.GetDirSize(path1)
	treeData = append(treeData, dto.CleanTree{ID: uuid.NewString(), Label: "rollback_app", Size: uint64(size1), Children: list1, Type: "rollback_app", IsRecommend: true})

	path2 := path.Join(global.Dir.BaseDir, rollbackPath, "website")
	list2 := loadTreeWithAllFile(true, path2, "rollback_website", path2, fileOp)
	size2, _ := fileOp.GetDirSize(path2)
	treeData = append(treeData, dto.CleanTree{ID: uuid.NewString(), Label: "rollback_website", Size: uint64(size2), Children: list2, Type: "rollback_website", IsRecommend: true})

	path3 := path.Join(global.Dir.BaseDir, rollbackPath, "database")
	list3 := loadTreeWithAllFile(true, path3, "rollback_database", path3, fileOp)
	size3, _ := fileOp.GetDirSize(path3)
	treeData = append(treeData, dto.CleanTree{ID: uuid.NewString(), Label: "rollback_database", Size: uint64(size3), Children: list3, Type: "rollback_database", IsRecommend: true})

	return treeData
}

func loadUnusedFile(fileOp fileUtils.FileOp) []dto.CleanTree {
	var treeData []dto.CleanTree
	path1 := path.Join(global.Dir.BaseDir, oldOriginalPath)
	list1 := loadTreeWithAllFile(true, path1, "old_original", path1, fileOp)
	if len(list1) != 0 {
		size, _ := fileOp.GetDirSize(path1)
		treeData = append(treeData, dto.CleanTree{ID: uuid.NewString(), Label: "old_original", Size: uint64(size), Children: list1, Type: "old_original"})
	}

	path2 := path.Join(global.Dir.BaseDir, oldAppBackupPath)
	list2 := loadTreeWithAllFile(true, path2, "old_apps_bak", path2, fileOp)
	if len(list2) != 0 {
		size, _ := fileOp.GetDirSize(path2)
		treeData = append(treeData, dto.CleanTree{ID: uuid.NewString(), Label: "old_apps_bak", Size: uint64(size), Children: list2, Type: "old_apps_bak"})
	}

	path3 := path.Join(global.Dir.BaseDir, oldDownloadPath)
	list3 := loadTreeWithAllFile(true, path3, "old_download", path3, fileOp)
	if len(list3) != 0 {
		size, _ := fileOp.GetDirSize(path3)
		treeData = append(treeData, dto.CleanTree{ID: uuid.NewString(), Label: "old_download", Size: uint64(size), Children: list3, Type: "old_download"})
	}

	path4 := path.Join(global.Dir.BaseDir, oldUpgradePath)
	list4 := loadTreeWithDir(true, "old_upgrade", path4, fileOp)
	itemSize := uint64(0)
	for _, item := range list4 {
		itemSize += item.Size
	}
	if len(list4) != 0 {
		treeData = append(treeData, dto.CleanTree{ID: uuid.NewString(), Label: "old_upgrade", Size: itemSize, Children: list4, Type: "old_upgrade"})
	}
	return treeData
}

func loadUploadTree(fileOp fileUtils.FileOp) []dto.CleanTree {
	var treeData []dto.CleanTree

	path0 := path.Join(global.Dir.BaseDir, tmpUploadPath)
	list0 := loadTreeWithAllFile(true, path0, "upload_tmp", path0, fileOp)
	size0, _ := fileOp.GetDirSize(path0)
	treeData = append(treeData, dto.CleanTree{ID: uuid.NewString(), Label: "upload_tmp", Size: uint64(size0), Children: list0, Type: "upload_tmp", IsRecommend: true})

	path1 := path.Join(global.Dir.BaseDir, uploadPath, "app")
	list1 := loadTreeWithAllFile(true, path1, "upload_app", path1, fileOp)
	size1, _ := fileOp.GetDirSize(path1)
	treeData = append(treeData, dto.CleanTree{ID: uuid.NewString(), Label: "upload_app", Size: uint64(size1), Children: list1, Type: "upload_app", IsRecommend: true})

	path2 := path.Join(global.Dir.BaseDir, uploadPath, "website")
	list2 := loadTreeWithAllFile(true, path2, "upload_website", path2, fileOp)
	size2, _ := fileOp.GetDirSize(path2)
	treeData = append(treeData, dto.CleanTree{ID: uuid.NewString(), Label: "upload_website", Size: uint64(size2), Children: list2, Type: "upload_website", IsRecommend: true})

	path3 := path.Join(global.Dir.BaseDir, uploadPath, "database")
	list3 := loadTreeWithAllFile(true, path3, "upload_database", path3, fileOp)
	size3, _ := fileOp.GetDirSize(path3)
	treeData = append(treeData, dto.CleanTree{ID: uuid.NewString(), Label: "upload_database", Size: uint64(size3), Children: list3, Type: "upload_database", IsRecommend: true})

	path4 := path.Join(global.Dir.BaseDir, uploadPath, "directory")
	list4 := loadTreeWithAllFile(true, path4, "upload_directory", path4, fileOp)
	size4, _ := fileOp.GetDirSize(path4)
	treeData = append(treeData, dto.CleanTree{ID: uuid.NewString(), Label: "upload_directory", Size: uint64(size4), Children: list4, Type: "upload_directory", IsRecommend: true})

	path5 := path.Join(global.Dir.BaseDir, uploadPath)
	uploadTreeData := loadTreeWithAllFile(true, path5, "upload", path5, fileOp)
	treeData = append(treeData, uploadTreeData...)

	return treeData
}

func loadDownloadTree(fileOp fileUtils.FileOp) []dto.CleanTree {
	var treeData []dto.CleanTree
	path1 := path.Join(global.Dir.BaseDir, downloadPath, "app")
	list1 := loadTreeWithAllFile(true, path1, "download_app", path1, fileOp)
	size1, _ := fileOp.GetDirSize(path1)
	treeData = append(treeData, dto.CleanTree{ID: uuid.NewString(), Label: "download_app", Size: uint64(size1), Children: list1, Type: "download_app", IsRecommend: true})

	path2 := path.Join(global.Dir.BaseDir, downloadPath, "website")
	list2 := loadTreeWithAllFile(true, path2, "download_website", path2, fileOp)
	size2, _ := fileOp.GetDirSize(path2)
	treeData = append(treeData, dto.CleanTree{ID: uuid.NewString(), Label: "download_website", Size: uint64(size2), Children: list2, Type: "download_website", IsRecommend: true})

	path3 := path.Join(global.Dir.BaseDir, downloadPath, "database")
	list3 := loadTreeWithAllFile(true, path3, "download_database", path3, fileOp)
	size3, _ := fileOp.GetDirSize(path3)
	treeData = append(treeData, dto.CleanTree{ID: uuid.NewString(), Label: "download_database", Size: uint64(size3), Children: list3, Type: "download_database", IsRecommend: true})

	path4 := path.Join(global.Dir.BaseDir, downloadPath, "directory")
	list4 := loadTreeWithAllFile(true, path4, "download_directory", path4, fileOp)
	size4, _ := fileOp.GetDirSize(path4)
	treeData = append(treeData, dto.CleanTree{ID: uuid.NewString(), Label: "download_directory", Size: uint64(size4), Children: list4, Type: "download_directory", IsRecommend: true})

	path5 := path.Join(global.Dir.BaseDir, downloadPath)
	uploadTreeData := loadTreeWithAllFile(true, path5, "download", path5, fileOp)
	treeData = append(treeData, uploadTreeData...)

	return treeData
}

func loadLogTree(fileOp fileUtils.FileOp) []dto.CleanTree {
	var treeData []dto.CleanTree
	path1 := path.Join(global.Dir.BaseDir, logPath)
	list1 := loadTreeWithAllFile(true, path1, "system_log", path1, fileOp)
	size := uint64(0)
	for _, file := range list1 {
		size += file.Size
	}
	treeData = append(treeData, dto.CleanTree{ID: uuid.NewString(), Label: "system_log", Size: uint64(size), Children: list1, Type: "system_log", IsRecommend: true})

	path2 := path.Join(global.Dir.BaseDir, logPath)
	list2 := loadTreeWithDir(false, "task_log", path2, fileOp)
	size2, _ := fileOp.GetDirSize(path2)
	treeData = append(treeData, dto.CleanTree{ID: uuid.NewString(), Label: "task_log", Size: uint64(size2), Children: list2, Type: "task_log"})
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
	treeData = append(treeData, dto.CleanTree{ID: uuid.NewString(), Label: "container_volumes", Size: volumeSize, Children: nil, Type: "volumes", IsRecommend: true})

	var buildCacheTotalSize int64
	for _, cache := range diskUsage.BuildCache {
		if cache.Type == "source.local" {
			buildCacheTotalSize += cache.Size
		}
	}
	treeData = append(treeData, dto.CleanTree{ID: uuid.NewString(), Label: "build_cache", Size: uint64(buildCacheTotalSize), Type: "build_cache", IsRecommend: true})
	return treeData
}

func loadTreeWithDir(isCheck bool, treeType, pathItem string, fileOp fileUtils.FileOp) []dto.CleanTree {
	var lists []dto.CleanTree
	files, err := os.ReadDir(pathItem)
	if err != nil {
		return lists
	}
	for _, file := range files {
		if treeType == "old_upgrade" {
			continue
		}
		if treeType == "task_log" && file.Name() == "ssl" {
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
	opts := types.BuildCachePruneOptions{}
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

func dropWithTask(itemPath string, taskItem *task.Task, size *int64, count *int) {
	itemSize := int64(0)
	itemCount := 0
	scanFile(itemPath, &itemSize, &itemCount)
	*size += itemSize
	*count += itemCount
	if err := os.RemoveAll(itemPath); err != nil {
		taskItem.LogFailed(fmt.Sprintf("drop file %s, err %v", itemPath, err))
		return
	}
	if itemCount != 0 {
		taskItem.LogSuccessF("drop file %s, size: %s, count: %d", itemPath, common.LoadSizeUnit2F(float64(itemSize)), itemCount)
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
