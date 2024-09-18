package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type snapHelper struct {
	SnapID      uint
	snapAgentDB *gorm.DB
	snapCoreDB  *gorm.DB
	Status      *model.SnapshotStatus
	Ctx         context.Context
	FileOp      files.FileOp
	Wg          *sync.WaitGroup
}

func loadDbConn(snap *snapHelper, targetDir string, req dto.SnapshotCreate) error {
	global.LOG.Debug("start load snapshot db conn")

	if err := snap.FileOp.CopyDir(path.Join(global.CONF.System.BaseDir, "1panel/db"), targetDir); err != nil {
		return err
	}
	agentDb, err := newSnapDB(path.Join(targetDir, "db"), "agent.db")
	if err != nil {
		return err
	}
	snap.snapAgentDB = agentDb
	coreDb, err := newSnapDB(path.Join(targetDir, "db"), "core.db")
	if err != nil {
		return err
	}
	snap.snapCoreDB = coreDb

	if !req.WithMonitorData {
		_ = os.Remove(path.Join(targetDir, "db/monitor.db"))
	}
	if !req.WithOperationLog {
		_ = snap.snapCoreDB.Exec("DELETE FROM operation_logs")
	}
	if !req.WithLoginLog {
		_ = snap.snapCoreDB.Exec("DELETE FROM login_logs")
	}
	if err := snap.snapAgentDB.Where("id = ?", snap.SnapID).Delete(&model.Snapshot{}).Error; err != nil {
		global.LOG.Errorf("delete current snapshot record failed, err: %v", err)
	}
	if err := snap.snapAgentDB.Where("snap_id = ?", snap.SnapID).Delete(&model.SnapshotStatus{}).Error; err != nil {
		global.LOG.Errorf("delete current snapshot status record failed, err: %v", err)
	}

	return nil
}

func snapBaseData(snap snapHelper, targetDir string) {
	defer snap.Wg.Done()
	_ = snapshotRepo.UpdateStatus(snap.Status.ID, map[string]interface{}{"base_data": constant.Running})
	status := constant.StatusDone
	if err := common.CopyFile("/usr/local/bin/1panel", targetDir); err != nil {
		status = err.Error()
	}
	if err := common.CopyFile("/usr/local/bin/1panel_agent", targetDir); err != nil {
		status = err.Error()
	}
	if err := common.CopyFile("/usr/local/bin/1pctl", targetDir); err != nil {
		status = err.Error()
	}
	if err := common.CopyFile("/etc/systemd/system/1panel.service", targetDir); err != nil {
		status = err.Error()
	}
	if err := common.CopyFile("/etc/systemd/system/1panel_agent.service", targetDir); err != nil {
		status = err.Error()
	}

	if snap.FileOp.Stat("/etc/docker/daemon.json") {
		if err := common.CopyFile("/etc/docker/daemon.json", targetDir); err != nil {
			status = err.Error()
		}
	}

	remarkInfo, _ := json.MarshalIndent(SnapshotJson{
		BaseDir:       global.CONF.System.BaseDir,
		BackupDataDir: global.CONF.System.Backup,
	}, "", "\t")
	if err := os.WriteFile(fmt.Sprintf("%s/snapshot.json", targetDir), remarkInfo, 0640); err != nil {
		status = err.Error()
	}
	snap.Status.BaseData = status
	_ = snapshotRepo.UpdateStatus(snap.Status.ID, map[string]interface{}{"base_data": status})
}

func snapAppImage(snap snapHelper, req dto.SnapshotCreate, targetDir string) {
	defer snap.Wg.Done()
	_ = snapshotRepo.UpdateStatus(snap.Status.ID, map[string]interface{}{"app_image": constant.Running})

	var imageList []string
	existStr, _ := cmd.Exec("docker images | awk '{print $1\":\"$2}' | grep -v REPOSITORY:TAG")
	existImages := strings.Split(existStr, "\n")
	for _, app := range req.AppData {
		for _, item := range app.Children {
			if item.Label == "appImage" && item.IsCheck {
				for _, existImage := range existImages {
					if existImage == item.Name {
						imageList = append(imageList, item.Name)
					}
				}
			}
		}
	}

	if len(imageList) != 0 {
		global.LOG.Debugf("docker save %s | gzip -c > %s", strings.Join(imageList, " "), path.Join(targetDir, "images.tar.gz"))
		std, err := cmd.Execf("docker save %s | gzip -c > %s", strings.Join(imageList, " "), path.Join(targetDir, "images.tar.gz"))
		if err != nil {
			snap.Status.AppImage = err.Error()
			_ = snapshotRepo.UpdateStatus(snap.Status.ID, map[string]interface{}{"app_image": std})
			return
		}
	}
	snap.Status.AppImage = constant.StatusDone
	_ = snapshotRepo.UpdateStatus(snap.Status.ID, map[string]interface{}{"app_image": constant.StatusDone})
}

func snapBackupData(snap snapHelper, req dto.SnapshotCreate, targetDir string) {
	defer snap.Wg.Done()
	_ = snapshotRepo.UpdateStatus(snap.Status.ID, map[string]interface{}{"backup_data": constant.Running})
	status := constant.StatusDone

	excludes := loadBackupExcludes(snap, req.BackupData)
	for _, item := range req.AppData {
		for _, itemApp := range item.Children {
			if itemApp.Label == "appBackup" {
				excludes = append(excludes, loadAppBackupExcludes([]dto.DataTree{itemApp})...)
			}
		}
	}
	global.LOG.Debugf("excludes backup file: %v\n", strings.Join(excludes, "\n"))

	if err := snap.FileOp.TarGzCompressPro(false, global.CONF.System.Backup, path.Join(targetDir, "1panel_backup.tar.gz"), "", strings.Join(excludes, ";")); err != nil {
		status = err.Error()
	}
	snap.Status.BackupData = status
	_ = snapshotRepo.UpdateStatus(snap.Status.ID, map[string]interface{}{"backup_data": status})
}
func loadBackupExcludes(snap snapHelper, req []dto.DataTree) []string {
	var excludes []string
	for _, item := range req {
		if len(item.Children) == 0 {
			if item.IsCheck {
				continue
			}
			if strings.HasPrefix(item.Path, path.Join(global.CONF.System.Backup, "system_snapshot")) {
				fmt.Println(strings.TrimSuffix(item.Name, ".tar.gz"))
				if err := snap.snapAgentDB.Debug().Where("name = ? AND download_account_id = ?", strings.TrimSuffix(item.Name, ".tar.gz"), "1").Delete(&model.Snapshot{}).Error; err != nil {
					global.LOG.Errorf("delete snapshot from database failed, err: %v", err)
				}
			} else {
				fmt.Println(strings.TrimPrefix(path.Dir(item.Path), global.CONF.System.Backup+"/"), path.Base(item.Path))
				if err := snap.snapAgentDB.Debug().Where("file_dir = ? AND file_name = ?", strings.TrimPrefix(path.Dir(item.Path), global.CONF.System.Backup+"/"), path.Base(item.Path)).Delete(&model.BackupRecord{}).Error; err != nil {
					global.LOG.Errorf("delete backup file from database failed, err: %v", err)
				}
			}
			fmt.Println(strings.TrimPrefix(item.Path, global.CONF.System.Backup))
			excludes = append(excludes, "."+strings.TrimPrefix(item.Path, global.CONF.System.Backup))
		} else {
			excludes = append(excludes, loadBackupExcludes(snap, item.Children)...)
		}
	}
	return excludes
}
func loadAppBackupExcludes(req []dto.DataTree) []string {
	var excludes []string
	for _, item := range req {
		if len(item.Children) == 0 {
			if !item.IsCheck {
				excludes = append(excludes, "."+strings.TrimPrefix(item.Path, path.Join(global.CONF.System.Backup)))
			}
		} else {
			excludes = append(excludes, loadAppBackupExcludes(item.Children)...)
		}
	}
	return excludes
}

func snapPanelData(snap snapHelper, req dto.SnapshotCreate, targetDir string) {
	defer snap.Wg.Done()
	_ = snapshotRepo.UpdateStatus(snap.Status.ID, map[string]interface{}{"panel_data": constant.Running})
	status := constant.StatusDone

	excludes := loadPanelExcludes(req.PanelData)
	for _, item := range req.AppData {
		for _, itemApp := range item.Children {
			if itemApp.Label == "appData" {
				excludes = append(excludes, loadPanelExcludes([]dto.DataTree{itemApp})...)
			}
		}
	}
	global.LOG.Debugf("excludes panel file: %v\n", strings.Join(excludes, "\n"))
	excludes = append(excludes, "./tmp")
	excludes = append(excludes, "./cache")
	excludes = append(excludes, "./uploads")
	excludes = append(excludes, "./db")
	excludes = append(excludes, "./resource")
	rootDir := path.Join(global.CONF.System.BaseDir, "1panel")
	if strings.Contains(global.CONF.System.Backup, rootDir) {
		excludes = append(excludes, "."+strings.ReplaceAll(global.CONF.System.Backup, rootDir, ""))
	}
	ignoreVal, _ := settingRepo.Get(settingRepo.WithByKey("SnapshotIgnore"))
	rules := strings.Split(ignoreVal.Value, ",")
	for _, ignore := range rules {
		if len(ignore) == 0 || cmd.CheckIllegal(ignore) {
			continue
		}
		excludes = append(excludes, "."+strings.ReplaceAll(ignore, rootDir, ""))
	}

	_ = snap.snapAgentDB.Model(&model.Setting{}).Where("key = ?", "SystemIP").Updates(map[string]interface{}{"value": ""})

	if err := snap.FileOp.TarGzCompressPro(false, rootDir, path.Join(targetDir, "1panel_data.tar.gz"), "", strings.Join(excludes, ";")); err != nil {
		status = err.Error()
	}

	snap.Status.PanelData = status
	_ = snapshotRepo.UpdateStatus(snap.Status.ID, map[string]interface{}{"panel_data": status})
}
func loadPanelExcludes(req []dto.DataTree) []string {
	var excludes []string
	for _, item := range req {
		if len(item.Children) == 0 {
			if !item.IsCheck {
				excludes = append(excludes, "."+strings.TrimPrefix(item.Path, path.Join(global.CONF.System.BaseDir, "1panel")))
			}
		} else {
			excludes = append(excludes, loadPanelExcludes(item.Children)...)
		}
	}
	return excludes
}

func snapCompress(snap snapHelper, rootDir string, secret string) {
	_ = snapshotRepo.UpdateStatus(snap.Status.ID, map[string]interface{}{"compress": constant.StatusRunning})
	tmpDir := path.Join(global.CONF.System.TmpDir, "system")
	fileName := fmt.Sprintf("%s.tar.gz", path.Base(rootDir))
	if err := snap.FileOp.TarGzCompressPro(true, rootDir, path.Join(tmpDir, fileName), secret, ""); err != nil {
		snap.Status.Compress = err.Error()
		_ = snapshotRepo.UpdateStatus(snap.Status.ID, map[string]interface{}{"compress": err.Error()})
		return
	}

	stat, err := os.Stat(path.Join(tmpDir, fileName))
	if err != nil {
		snap.Status.Compress = err.Error()
		_ = snapshotRepo.UpdateStatus(snap.Status.ID, map[string]interface{}{"compress": err.Error()})
		return
	}
	size := common.LoadSizeUnit2F(float64(stat.Size()))
	global.LOG.Debugf("compress successful! size of file: %s", size)
	snap.Status.Compress = constant.StatusDone
	snap.Status.Size = size
	_ = snapshotRepo.UpdateStatus(snap.Status.ID, map[string]interface{}{"compress": constant.StatusDone, "size": size})

	global.LOG.Debugf("remove snapshot file %s", rootDir)
	_ = os.RemoveAll(rootDir)
}

func snapUpload(snap snapHelper, accounts string, file string) {
	source := path.Join(global.CONF.System.TmpDir, "system", path.Base(file))
	_ = snapshotRepo.UpdateStatus(snap.Status.ID, map[string]interface{}{"upload": constant.StatusUploading})
	accountMap, err := NewBackupClientMap(strings.Split(accounts, ","))
	if err != nil {
		snap.Status.Upload = err.Error()
		_ = snapshotRepo.UpdateStatus(snap.Status.ID, map[string]interface{}{"upload": err.Error()})
		return
	}
	targetAccounts := strings.Split(accounts, ",")
	for _, item := range targetAccounts {
		global.LOG.Debugf("start upload snapshot to %s, path: %s", item, path.Join(accountMap[item].backupPath, "system_snapshot", path.Base(file)))
		if _, err := accountMap[item].client.Upload(source, path.Join(accountMap[item].backupPath, "system_snapshot", path.Base(file))); err != nil {
			global.LOG.Debugf("upload to %s failed, err: %v", accountMap[item].name, err)
			snap.Status.Upload = err.Error()
			_ = snapshotRepo.UpdateStatus(snap.Status.ID, map[string]interface{}{"upload": err.Error()})
			return
		}
		global.LOG.Debugf("upload to %s successful", accountMap[item].name)
	}
	snap.Status.Upload = constant.StatusDone
	_ = snapshotRepo.UpdateStatus(snap.Status.ID, map[string]interface{}{"upload": constant.StatusDone})

	global.LOG.Debugf("remove snapshot file %s", source)
	_ = os.Remove(source)
}

func newSnapDB(dir, file string) (*gorm.DB, error) {
	db, _ := gorm.Open(sqlite.Open(path.Join(dir, file)), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetConnMaxIdleTime(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	// global.LOG.Debug("load snapshot db conn successful!")
	return db, nil
}

func closeDatabase(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		return
	}
	_ = sqlDB.Close()
}
