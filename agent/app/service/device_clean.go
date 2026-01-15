package service

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	fileUtils "github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/build"
	"github.com/docker/docker/api/types/filters"
	"github.com/google/uuid"
)

const (
	rollbackPath = "1panel/tmp"
	upgradePath  = "1panel/tmp/upgrade"
	uploadPath   = "1panel/uploads"
	downloadPath = "1panel/download"
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
	treeData = append(treeData, loadUpgradeTree(fileOp))
	treeData = append(treeData, loadAgentPackage(fileOp))

	SystemClean.BackupClean = loadBackupTree(fileOp)

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
	for _, item := range req {
		size += item.Size
		switch item.TreeType {
		case "1panel_original":
			dropFileOrDir(path.Join(global.Dir.BaseDir, "1panel_original", item.Name))

		case "upgrade":
			dropFileOrDir(path.Join(global.Dir.BaseDir, upgradePath, item.Name))

		case "agent":
			dropFileOrDir(path.Join(global.Dir.BaseDir, "1panel/agent/package", item.Name))

		case "tmp_backup":
			dropFileOrDir(path.Join(global.Dir.LocalBackupDir, "tmp"))
		case "unknown_backup":
			if strings.HasPrefix(item.Name, path.Join(global.Dir.LocalBackupDir, "log/website")) {
				dropFileOrDir(item.Name)
			} else {
				dropFile(item.Name)
			}

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

		case "upload":
			dropFileOrDir(path.Join(global.Dir.BaseDir, uploadPath, item.Name))
		case "upload_app":
			dropFileOrDir(path.Join(global.Dir.BaseDir, uploadPath, "app", item.Name))
		case "upload_database":
			dropFileOrDir(path.Join(global.Dir.BaseDir, uploadPath, "database", item.Name))
		case "upload_website":
			dropFileOrDir(path.Join(global.Dir.BaseDir, uploadPath, "website", item.Name))
		case "download":
			dropFileOrDir(path.Join(global.Dir.BaseDir, downloadPath, item.Name))
		case "download_app":
			dropFileOrDir(path.Join(global.Dir.BaseDir, downloadPath, "app", item.Name))
		case "download_database":
			dropFileOrDir(path.Join(global.Dir.BaseDir, downloadPath, "database", item.Name))
		case "download_website":
			dropFileOrDir(path.Join(global.Dir.BaseDir, downloadPath, "website", item.Name))

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
				files, _ := os.ReadDir(global.Dir.TaskDir)
				if len(files) == 0 {
					continue
				}
				for _, file := range files {
					if file.Name() == "ssl" || !file.IsDir() {
						continue
					}
					dropTaskLog(path.Join(global.Dir.TaskDir, file.Name()))
				}
			} else {
				dropTaskLog(path.Join(global.Dir.TaskDir, item.Name))
			}
		case "website_log":
			dropWebsiteLog(item.Name)
		case "script":
			dropFileOrDir(path.Join(global.Dir.TmpDir, "script", item.Name))
		case "images":
			_, _ = dropImages()
		case "containers":
			_, _ = dropContainers()
		case "volumes":
			_, _ = dropVolumes()
		case "build_cache":
			_, _ = dropBuildCache()
		case "app_tmp_download_version":
			dropFileOrDir(path.Join(global.Dir.RemoteAppResourceDir, item.Name))
		}
	}

	_ = cleanEmptyDirs(global.Dir.LocalBackupDir)
	_ = settingRepo.Update("LastCleanTime", time.Now().Format(constant.DateTimeLayout))
	_ = settingRepo.Update("LastCleanSize", fmt.Sprintf("%v", size))
	_ = settingRepo.Update("LastCleanData", fmt.Sprintf("%v", len(req)))
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

		dropWithTask(path.Join(global.Dir.LocalBackupDir, "tmp/system"), taskItem, &size, &fileCount)

		dropWithTask(path.Join(global.Dir.BaseDir, rollbackPath, "app"), taskItem, &size, &fileCount)
		dropWithTask(path.Join(global.Dir.BaseDir, rollbackPath, "website"), taskItem, &size, &fileCount)
		dropWithTask(path.Join(global.Dir.BaseDir, rollbackPath, "database"), taskItem, &size, &fileCount)

		upgrades := path.Join(global.Dir.BaseDir, upgradePath)
		oldUpgradeFiles, _ := os.ReadDir(upgrades)
		if len(oldUpgradeFiles) != 0 {
			for i := 0; i < len(oldUpgradeFiles); i++ {
				dropWithTask(path.Join(upgrades, oldUpgradeFiles[i].Name()), taskItem, &size, &fileCount)
			}
		}

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

		count1, size1 := dropVolumes()
		size += int64(size1)
		fileCount += count1
		count2, size2 := dropBuildCache()
		size += int64(size2)
		fileCount += count2

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

func loadUpgradeTree(fileOp fileUtils.FileOp) dto.CleanTree {
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
		if global.IsMaster {
			var copiesSetting model.Setting
			_ = global.CoreDB.Where("key = ?", "UpgradeBackupCopies").First(&copiesSetting).Error
			copies, _ := strconv.Atoi(copiesSetting.Value)
			if copies == 0 || copies > len(upgradeTree.Children) {
				copies = len(upgradeTree.Children)
			}
			for i := 0; i < copies; i++ {
				upgradeTree.Children[i].IsCheck = false
				upgradeTree.Children[i].IsRecommend = false
			}
		} else {
			upgradeTree.Children[0].IsCheck = false
			upgradeTree.Children[0].IsRecommend = false
		}
	}
	return upgradeTree
}

func loadAgentPackage(fileOp fileUtils.FileOp) dto.CleanTree {
	pathItem := path.Join(global.Dir.BaseDir, "1panel/agent/package")
	itemTree := dto.CleanTree{
		ID:          uuid.NewString(),
		Label:       "agent_packages",
		IsCheck:     false,
		IsRecommend: true,
		Type:        "agent",
	}
	files, _ := os.ReadDir(pathItem)
	for _, file := range files {
		if file.IsDir() {
			itemSize, _ := fileOp.GetDirSize(path.Join(pathItem, file.Name()))
			itemTree.Size += uint64(itemSize)
			itemTree.Children = append(itemTree.Children, dto.CleanTree{
				ID:          uuid.NewString(),
				Label:       file.Name(),
				Name:        file.Name(),
				Size:        uint64(itemSize),
				IsCheck:     true,
				IsRecommend: true,
				Type:        "agent",
			})
		} else {
			itemSize, _ := file.Info()
			itemName := file.Name()
			isCurrentVersion := strings.HasPrefix(itemName, fmt.Sprintf("1panel-agent_%s_", global.CONF.Base.Version))
			if isCurrentVersion {
				continue
			}
			itemTree.Size += uint64(itemSize.Size())
			itemTree.Children = append(itemTree.Children, dto.CleanTree{
				ID:          uuid.NewString(),
				Label:       itemName,
				Name:        itemName,
				Size:        uint64(itemSize.Size()),
				IsCheck:     !isCurrentVersion,
				IsRecommend: true,
				Type:        "agent",
			})
		}
	}
	if itemTree.Size == 0 {
		itemTree.IsCheck = false
	}
	return itemTree
}

func loadBackupTree(fileOp fileUtils.FileOp) []dto.CleanTree {
	var treeData []dto.CleanTree

	tmpSize, _ := fileOp.GetDirSize(path.Join(global.Dir.LocalBackupDir, "tmp"))
	treeData = append(treeData, dto.CleanTree{
		ID:          uuid.NewString(),
		Label:       "tmp_backup",
		Size:        uint64(tmpSize),
		IsCheck:     tmpSize != 0,
		IsRecommend: true,
		Type:        "tmp_backup",
	})
	backupRecords, _ := backupRepo.ListRecord()
	var recordMap = make(map[string][]string)
	for _, record := range backupRecords {
		if val, ok := recordMap[record.FileDir]; ok {
			val = append(val, record.FileName)
			recordMap[record.FileDir] = val
		} else {
			recordMap[record.FileDir] = []string{record.FileName}
		}
	}

	treeData = append(treeData, loadUnknownApps(fileOp, recordMap))
	treeData = append(treeData, loadUnknownDbs(fileOp, recordMap))
	treeData = append(treeData, loadUnknownWebsites(fileOp, recordMap))
	treeData = append(treeData, loadUnknownSnapshot(fileOp))
	treeData = append(treeData, loadUnknownWebsiteLog(fileOp))
	return treeData
}

func loadUnknownApps(fileOp fileUtils.FileOp, recordMap map[string][]string) dto.CleanTree {
	apps, _ := appInstallRepo.ListBy(context.Background())
	var excludePaths []string
	for _, app := range apps {
		itemName := fmt.Sprintf("app/%s/%s", app.App.Key, app.Name)
		if val, ok := recordMap[itemName]; ok {
			for _, item := range val {
				excludePaths = append(excludePaths, path.Join(global.Dir.LocalBackupDir, itemName, item))
			}
		}
	}
	backupPath := path.Join(global.Dir.LocalBackupDir, "app")
	treeData := dto.CleanTree{
		ID:          uuid.NewString(),
		Label:       "unknown_app",
		IsCheck:     false,
		IsRecommend: false,
		Name:        backupPath,
		Type:        "unknown_backup",
	}
	_ = loadFileOrDirWithExclude(fileOp, 0, backupPath, &treeData, excludePaths)
	return treeData
}
func loadUnknownDbs(fileOp fileUtils.FileOp, recordMap map[string][]string) dto.CleanTree {
	dbs, _ := databaseRepo.GetList()
	var excludePaths []string
	dbMap := make(map[string]struct{})
	for _, db := range dbs {
		dbMap[fmt.Sprintf("database/%s/%s", db.Type, db.Name)] = struct{}{}
	}
	for key, val := range recordMap {
		itemName := path.Dir(key)
		if _, ok := dbMap[itemName]; ok {
			for _, item := range val {
				excludePaths = append(excludePaths, path.Join(global.Dir.LocalBackupDir, key, item))
			}
		}
	}
	backupPath := path.Join(global.Dir.LocalBackupDir, "database")
	treeData := dto.CleanTree{
		ID:          uuid.NewString(),
		Label:       "unknown_database",
		Name:        backupPath,
		IsCheck:     false,
		IsRecommend: false,
		Type:        "unknown_backup",
	}
	_ = loadFileOrDirWithExclude(fileOp, 0, backupPath, &treeData, excludePaths)
	return treeData
}
func loadUnknownWebsites(fileOp fileUtils.FileOp, recordMap map[string][]string) dto.CleanTree {
	websites, _ := websiteRepo.List()
	var excludePaths []string
	for _, website := range websites {
		itemName := fmt.Sprintf("website/%s", website.Alias)
		if val, ok := recordMap[itemName]; ok {
			for _, item := range val {
				excludePaths = append(excludePaths, path.Join(global.Dir.LocalBackupDir, itemName, item))
			}
		}
	}
	backupPath := path.Join(global.Dir.LocalBackupDir, "website")
	treeData := dto.CleanTree{
		ID:          uuid.NewString(),
		Label:       "unknown_website",
		Name:        backupPath,
		IsCheck:     false,
		IsRecommend: false,
		Type:        "unknown_backup",
	}
	_ = loadFileOrDirWithExclude(fileOp, 0, backupPath, &treeData, excludePaths)
	return treeData
}
func loadUnknownSnapshot(fileOp fileUtils.FileOp) dto.CleanTree {
	snaps, _ := snapshotRepo.GetList()
	var excludePaths []string
	for _, item := range snaps {
		excludePaths = append(excludePaths, path.Join(global.Dir.LocalBackupDir, "system_snapshot", item.Name+".tar.gz"))
	}
	backupPath := path.Join(global.Dir.LocalBackupDir, "system_snapshot")
	treeData := dto.CleanTree{
		ID:          uuid.NewString(),
		Label:       "unknown_snapshot",
		Name:        backupPath,
		IsCheck:     false,
		IsRecommend: false,
		Type:        "unknown_backup",
	}
	entries, _ := os.ReadDir(backupPath)
	for _, entry := range entries {
		childPath := filepath.Join(backupPath, entry.Name())
		if isExactPathMatch(childPath, excludePaths) {
			continue
		}
		childNode := dto.CleanTree{
			ID:          uuid.NewString(),
			Label:       entry.Name(),
			IsCheck:     false,
			IsRecommend: false,
			Name:        childPath,
			Type:        "unknown_backup",
		}
		if entry.IsDir() {
			itemSize, _ := fileOp.GetDirSize(childPath)
			childNode.Size = uint64(itemSize)
			childNode.IsCheck = true
			childNode.IsRecommend = true
			treeData.Size += childNode.Size
		} else {
			info, _ := entry.Info()
			childNode.Size = uint64(info.Size())
			treeData.Size += childNode.Size
		}

		treeData.Children = append(treeData.Children, childNode)
	}
	return treeData
}

func loadUnknownWebsiteLog(fileOp fileUtils.FileOp) dto.CleanTree {
	treeData := dto.CleanTree{
		ID:          uuid.NewString(),
		Label:       "unknown_website_log",
		IsCheck:     false,
		IsRecommend: true,
		Type:        "unknown_backup",
	}
	dir := path.Join(global.Dir.LocalBackupDir, "log/website")
	websites, _ := websiteRepo.List()
	websiteMap := make(map[string]struct{})
	for _, website := range websites {
		websiteMap[website.Alias] = struct{}{}
	}

	entries, _ := os.ReadDir(dir)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dirName := entry.Name()
		if _, ok := websiteMap[dirName]; !ok {
			dirPath := path.Join(dir, dirName)
			itemSize, _ := fileOp.GetDirSize(dirPath)
			childData := dto.CleanTree{
				ID:          uuid.NewString(),
				Label:       dirName,
				IsCheck:     true,
				IsRecommend: true,
				Name:        dirPath,
				Type:        "unknown_backup",
				Size:        uint64(itemSize),
			}
			treeData.Size += uint64(itemSize)
			treeData.Children = append(treeData.Children, childData)
		}
	}
	if treeData.Size > 0 {
		treeData.IsCheck = true
	}
	return treeData
}

func loadFileOrDirWithExclude(fileOp fileUtils.FileOp, index uint, dir string, rootTree *dto.CleanTree, excludes []string) error {
	index++
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		childPath := filepath.Join(dir, entry.Name())
		if isExactPathMatch(childPath, excludes) {
			continue
		}
		childNode := dto.CleanTree{
			ID:          uuid.NewString(),
			Label:       entry.Name(),
			IsCheck:     false,
			IsRecommend: false,
			Name:        childPath,
			Type:        "unknown_backup",
		}
		if entry.IsDir() {
			if index < 4 {
				if err = loadFileOrDirWithExclude(fileOp, index, childPath, &childNode, excludes); err != nil {
					return err
				}
				childNode.Size = 0
				for _, child := range childNode.Children {
					childNode.Size += child.Size
				}
				rootTree.Size += childNode.Size
			} else {
				itemSize, _ := fileOp.GetDirSize(childPath)
				childNode.Size = uint64(itemSize)
				rootTree.Size += childNode.Size
			}
		} else {
			info, _ := entry.Info()
			childNode.Size = uint64(info.Size())
			rootTree.Size += childNode.Size
		}

		rootTree.Children = append(rootTree.Children, childNode)
	}
	return nil
}

func isExactPathMatch(path string, excludePaths []string) bool {
	cleanPath := filepath.Clean(path)

	for _, excludePath := range excludePaths {
		cleanExclude := filepath.Clean(excludePath)
		if cleanPath == cleanExclude {
			return true
		}
	}

	return false
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

	appTmpDownloadTree := loadAppTmpDownloadTree(fileOp)
	if len(appTmpDownloadTree) > 0 {
		parentTree := dto.CleanTree{
			ID:          uuid.NewString(),
			Label:       "app_tmp_download",
			IsCheck:     true,
			IsRecommend: true,
			Type:        "app_tmp_download",
			Name:        "apps",
		}
		for _, child := range appTmpDownloadTree {
			parentTree.Size += child.Size
		}
		parentTree.Children = appTmpDownloadTree
		treeData = append(treeData, parentTree)
	}
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
	treeData = append(treeData, dto.CleanTree{ID: uuid.NewString(), Label: "system_log", Size: size, Children: list1, Type: "system_log", IsRecommend: true})

	path2 := path.Join(global.Dir.TaskDir)
	list2 := loadTreeWithDir(false, "task_log", path2, fileOp)
	size2, _ := fileOp.GetDirSize(path2)
	treeData = append(treeData, dto.CleanTree{ID: uuid.NewString(), Label: "task_log", Size: uint64(size2), Children: list2, Type: "task_log"})

	websiteLogList := loadWebsiteLogTree(fileOp)
	logTotalSize := uint64(0)
	for _, websiteLog := range websiteLogList {
		logTotalSize += websiteLog.Size
	}
	treeData = append(treeData, dto.CleanTree{ID: uuid.NewString(), Label: "website_log", Size: logTotalSize, Children: websiteLogList, Type: "website_log", IsRecommend: false})

	return treeData
}

func loadWebsiteLogTree(fileOp fileUtils.FileOp) []dto.CleanTree {
	websites, _ := websiteRepo.List()
	if len(websites) == 0 {
		return nil
	}
	var res []dto.CleanTree
	for _, website := range websites {
		size3, _ := fileOp.GetDirSize(path.Join(GetSiteDir(website.Alias), "log"))
		res = append(res, dto.CleanTree{
			ID:    uuid.NewString(),
			Label: website.PrimaryDomain,
			Size:  uint64(size3),
			Type:  "website_log",
			Name:  website.Alias,
		})
	}
	return res
}

func loadAppTmpDownloadTree(fileOp fileUtils.FileOp) []dto.CleanTree {
	appDirs, err := os.ReadDir(global.Dir.RemoteAppResourceDir)
	if err != nil {
		return nil
	}
	var res []dto.CleanTree
	for _, appDir := range appDirs {
		if !appDir.IsDir() {
			continue
		}
		appKey := appDir.Name()
		app, _ := appRepo.GetFirst(appRepo.WithKey(appKey))
		if app.ID == 0 {
			continue
		}
		appPath := filepath.Join(global.Dir.RemoteAppResourceDir, appKey)
		versionDirs, err := os.ReadDir(appPath)
		if err != nil {
			continue
		}
		appDetails, _ := appDetailRepo.GetBy(appDetailRepo.WithAppId(app.ID))
		existingVersions := make(map[string]bool)
		for _, appDetail := range appDetails {
			existingVersions[appDetail.Version] = true
		}
		var missingVersions []string
		for _, versionDir := range versionDirs {
			if !versionDir.IsDir() {
				continue
			}

			version := versionDir.Name()
			if !existingVersions[version] {
				missingVersions = append(missingVersions, version)
			}
		}
		if len(missingVersions) > 0 {
			var appTree dto.CleanTree
			appTree.ID = uuid.NewString()
			appTree.Label = app.Name
			appTree.Type = "app_tmp_download"
			appTree.Name = appKey
			appTree.IsRecommend = true
			appTree.IsCheck = true
			for _, version := range missingVersions {
				versionPath := filepath.Join(appPath, version)
				size, _ := fileOp.GetDirSize(versionPath)
				appTree.Size += uint64(size)
				appTree.Children = append(appTree.Children, dto.CleanTree{
					ID:          uuid.NewString(),
					Label:       version,
					Size:        uint64(size),
					IsCheck:     true,
					IsRecommend: true,
					Type:        "app_tmp_download_version",
					Name:        path.Join(appKey, version),
				})
			}
			res = append(res, appTree)
		}
	}
	return res
}

func loadContainerTree() []dto.CleanTree {
	var treeData []dto.CleanTree
	client, err := docker.NewDockerClient()
	if err != nil {
		return treeData
	}
	defer client.Close()
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
			Size:        size,
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
	if err := os.RemoveAll(itemPath); err != nil {
		global.LOG.Errorf("drop file %s failed, err %v", itemPath, err)
	}
}
func dropFile(itemPath string) {
	info, err := os.Stat(itemPath)
	if err != nil {
		return
	}
	if info.IsDir() {
		return
	}
	if err := os.Remove(itemPath); err != nil {
		global.LOG.Errorf("drop file %s failed, err %v", itemPath, err)
	}
}

func dropBuildCache() (int, int) {
	client, err := docker.NewDockerClient()
	if err != nil {
		global.LOG.Errorf("do not get docker client")
		return 0, 0
	}
	defer client.Close()
	opts := build.CachePruneOptions{}
	opts.All = true
	res, err := client.BuildCachePrune(context.Background(), opts)
	if err != nil {
		global.LOG.Errorf("drop build cache failed, err %v", err)
		return 0, 0
	}
	return len(res.CachesDeleted), int(res.SpaceReclaimed)
}

func dropImages() (int, int) {
	client, err := docker.NewDockerClient()
	if err != nil {
		global.LOG.Errorf("do not get docker client")
		return 0, 0
	}
	defer client.Close()
	pruneFilters := filters.NewArgs()
	pruneFilters.Add("dangling", "false")
	res, err := client.ImagesPrune(context.Background(), pruneFilters)
	if err != nil {
		global.LOG.Errorf("drop images failed, err %v", err)
		return 0, 0
	}
	return len(res.ImagesDeleted), int(res.SpaceReclaimed)
}

func dropContainers() (int, int) {
	client, err := docker.NewDockerClient()
	if err != nil {
		global.LOG.Errorf("do not get docker client")
		return 0, 0
	}
	defer client.Close()
	pruneFilters := filters.NewArgs()
	res, err := client.ContainersPrune(context.Background(), pruneFilters)
	if err != nil {
		global.LOG.Errorf("drop containers failed, err %v", err)
		return 0, 0
	}
	return len(res.ContainersDeleted), int(res.SpaceReclaimed)
}

func dropVolumes() (int, int) {
	client, err := docker.NewDockerClient()
	if err != nil {
		global.LOG.Errorf("do not get docker client")
		return 0, 0
	}
	defer client.Close()
	pruneFilters := filters.NewArgs()
	versions, err := client.ServerVersion(context.Background())
	if err != nil {
		global.LOG.Errorf("do not get docker api versions")
		return 0, 0
	}
	if common.ComparePanelVersion(versions.APIVersion, "1.42") {
		pruneFilters.Add("all", "true")
	}
	res, err := client.VolumesPrune(context.Background(), pruneFilters)
	if err != nil {
		global.LOG.Errorf("drop volumes failed, err %v", err)
		return 0, 0
	}
	return len(res.VolumesDeleted), int(res.SpaceReclaimed)
}

func dropWebsiteLog(alias string) {
	accessLogPath := path.Join(GetSiteDir(alias), "log", "access.log")
	errorLogPath := path.Join(GetSiteDir(alias), "log", "error.log")
	if err := os.Truncate(accessLogPath, 0); err != nil {
		global.LOG.Errorf("truncate access log %s failed, err %v", accessLogPath, err)
	}

	if err := os.Truncate(errorLogPath, 0); err != nil {
		global.LOG.Errorf("truncate error log %s failed, err %v", errorLogPath, err)
	}
}

func dropTaskLog(logDir string) {
	files, err := os.ReadDir(logDir)
	if err != nil {
		return
	}
	taskType := path.Base(logDir)
	var usedTasks []string
	switch taskType {
	case "Cronjob":
		_ = global.DB.Model(&model.JobRecords{}).Where("task_id != ?", "").Select("task_id").Find(&usedTasks).Error
	case "Snapshot":
		var (
			snapIDs     []string
			recoverIDs  []string
			rollbackIDs []string
		)
		_ = global.DB.Model(&model.Snapshot{}).Where("task_id != ?", "").Select("task_id").Find(&snapIDs).Error
		_ = global.DB.Model(&model.Snapshot{}).Where("task_recover_id != ", "").Select("task_id").Find(&recoverIDs).Error
		_ = global.DB.Model(&model.Snapshot{}).Where("task_rollback_id != ?", "").Select("task_id").Find(&rollbackIDs).Error
		usedTasks = append(usedTasks, snapIDs...)
		usedTasks = append(usedTasks, recoverIDs...)
		usedTasks = append(usedTasks, rollbackIDs...)
	case "Backup":
		_ = global.DB.Model(&model.BackupRecord{}).Where("task_id != ?", "").Select("task_id").Find(&usedTasks).Error
	case "Clam":
		_ = global.DB.Model(&model.ClamRecord{}).Where("task_id != ?", "").Select("task_id").Find(&usedTasks).Error
	case "Tamper":
		xpackDB, err := common.LoadDBConnByPathWithErr(path.Join(global.CONF.Base.InstallDir, "1panel/db/xpack.db"), "xpack.db")
		if err == nil {
			_ = xpackDB.Table("tampers").Where("task_id != ?", "").Select("task_id").Find(&usedTasks).Error
		}
		defer common.CloseDB(xpackDB)
	case "System":
		xpackDB, err := common.LoadDBConnByPathWithErr(path.Join(global.CONF.Base.InstallDir, "1panel/db/xpack.db"), "xpack.db")
		if err == nil {
			_ = xpackDB.Model("nodes").Where("task_id != ?", "").Select("task_id").Find(&usedTasks).Error
		}
		defer common.CloseDB(xpackDB)
	default:
		dropFileOrDir(logDir)
		_ = taskRepo.Delete(repo.WithByType(taskType))
		return
	}
	usedMap := make(map[string]struct{})
	for _, item := range usedTasks {
		if _, ok := usedMap[item]; !ok {
			usedMap[item] = struct{}{}
		}
	}
	for _, item := range files {
		if _, ok := usedMap[strings.TrimSuffix(item.Name(), ".log")]; ok {
			continue
		}
		_ = os.Remove(logDir + "/" + item.Name())
	}
	_ = taskRepo.Delete(repo.WithByType(taskType), taskRepo.WithByIDNotIn(usedTasks))
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
		taskItem.Log(i18n.GetWithNameAndErr("FileDropFailed", itemPath, err))
		return
	}
	if itemCount != 0 {
		taskItem.Log(i18n.GetMsgWithMap("FileDropSuccess", map[string]interface{}{"name": itemPath, "count": itemCount, "size": common.LoadSizeUnit2F(float64(itemSize))}))
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

func cleanEmptyDirs(root string) error {
	dirsToCheck := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			dirsToCheck = append(dirsToCheck, path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	for i := len(dirsToCheck) - 1; i >= 0; i-- {
		dir := dirsToCheck[i]
		if dir == root {
			continue
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		if len(entries) == 0 {
			_ = os.Remove(dir)
		}
	}
	return nil
}
