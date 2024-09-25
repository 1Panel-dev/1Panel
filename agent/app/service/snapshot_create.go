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
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/compose"
	"github.com/1Panel-dev/1Panel/agent/utils/copier"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/glebarez/sqlite"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func (u *SnapshotService) SnapshotCreate(req dto.SnapshotCreate) error {
	versionItem, _ := settingRepo.Get(settingRepo.WithByKey("SystemVersion"))

	req.Name = fmt.Sprintf("1panel-%s-linux-%s-%s", versionItem.Value, loadOs(), time.Now().Format(constant.DateTimeSlimLayout))
	appItem, _ := json.Marshal(req.AppData)
	panelItem, _ := json.Marshal(req.PanelData)
	backupItem, _ := json.Marshal(req.BackupData)
	snap := model.Snapshot{
		Name:              req.Name,
		TaskID:            req.TaskID,
		Secret:            req.Secret,
		Description:       req.Description,
		SourceAccountIDs:  req.SourceAccountIDs,
		DownloadAccountID: req.DownloadAccountID,

		AppData:          string(appItem),
		PanelData:        string(panelItem),
		BackupData:       string(backupItem),
		WithMonitorData:  req.WithMonitorData,
		WithLoginLog:     req.WithLoginLog,
		WithOperationLog: req.WithOperationLog,
		WithTaskLog:      req.WithTaskLog,
		WithSystemLog:    req.WithSystemLog,

		Version: versionItem.Value,
		Status:  constant.StatusWaiting,
	}
	if err := snapshotRepo.Create(&snap); err != nil {
		global.LOG.Errorf("create snapshot record to db failed, err: %v", err)
		return err
	}

	req.ID = snap.ID
	if err := u.HandleSnapshot(req); err != nil {
		return err
	}
	return nil
}

func (u *SnapshotService) SnapshotReCreate(id uint) error {
	snap, err := snapshotRepo.Get(commonRepo.WithByID(id))
	if err != nil {
		return err
	}
	taskModel, err := taskRepo.GetFirst(taskRepo.WithResourceID(snap.ID), commonRepo.WithByType(task.TaskScopeSnapshot))
	if err != nil {
		return err
	}

	var req dto.SnapshotCreate
	_ = copier.Copy(&req, snap)
	if err := json.Unmarshal([]byte(snap.PanelData), &req.PanelData); err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(snap.AppData), &req.AppData); err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(snap.BackupData), &req.BackupData); err != nil {
		return err
	}
	req.TaskID = taskModel.ID
	if err := u.HandleSnapshot(req); err != nil {
		return err
	}

	return nil
}

func (u *SnapshotService) HandleSnapshot(req dto.SnapshotCreate) error {
	taskItem, err := task.NewTaskWithOps(req.Name, task.TaskCreate, task.TaskScopeSnapshot, req.TaskID, req.ID)
	if err != nil {
		global.LOG.Errorf("new task for create snapshot failed, err: %v", err)
		return err
	}

	rootDir := path.Join(global.CONF.System.BaseDir, "1panel/tmp/system", req.Name)
	itemHelper := snapHelper{SnapID: req.ID, Task: *taskItem, FileOp: files.NewFileOp(), Ctx: context.Background()}
	baseDir := path.Join(rootDir, "base")
	_ = os.MkdirAll(baseDir, os.ModePerm)

	go func() {
		taskItem.AddSubTask(
			i18n.GetMsgByKey("SnapDBInfo"),
			func(t *task.Task) error { return loadDbConn(&itemHelper, rootDir, req) },
			nil,
		)

		if len(req.InterruptStep) == 0 || req.InterruptStep == "SnapBaseInfo" {
			taskItem.AddSubTask(
				i18n.GetMsgByKey("SnapBaseInfo"),
				func(t *task.Task) error { return snapBaseData(itemHelper, baseDir) },
				nil,
			)
			req.InterruptStep = ""
		}
		if len(req.InterruptStep) == 0 || req.InterruptStep == "SnapInstallApp" {
			taskItem.AddSubTask(
				i18n.GetMsgByKey("SnapInstallApp"),
				func(t *task.Task) error { return snapAppImage(itemHelper, req, rootDir) },
				nil,
			)
			req.InterruptStep = ""
		}
		if len(req.InterruptStep) == 0 || req.InterruptStep == "SnapLocalBackup" {
			taskItem.AddSubTask(
				i18n.GetMsgByKey("SnapLocalBackup"),
				func(t *task.Task) error { return snapBackupData(itemHelper, req, rootDir) },
				nil,
			)
			req.InterruptStep = ""
		}
		if len(req.InterruptStep) == 0 || req.InterruptStep == "SnapPanelData" {
			taskItem.AddSubTask(
				i18n.GetMsgByKey("SnapPanelData"),
				func(t *task.Task) error { return snapPanelData(itemHelper, req, rootDir) },
				nil,
			)
			req.InterruptStep = ""
		}

		taskItem.AddSubTask(
			i18n.GetMsgByKey("SnapCloseDBConn"),
			func(t *task.Task) error {
				taskItem.Log("<######################## 6 / 8 ########################>")
				closeDatabase(itemHelper.snapAgentDB)
				closeDatabase(itemHelper.snapCoreDB)
				return nil
			},
			nil,
		)
		if len(req.InterruptStep) == 0 || req.InterruptStep == "SnapCompress" {
			taskItem.AddSubTask(
				i18n.GetMsgByKey("SnapCompress"),
				func(t *task.Task) error { return snapCompress(itemHelper, rootDir, req.Secret) },
				nil,
			)
			req.InterruptStep = ""
		}
		if len(req.InterruptStep) == 0 || req.InterruptStep == "SnapUpload" {
			taskItem.AddSubTask(
				i18n.GetMsgByKey("SnapUpload"),
				func(t *task.Task) error {
					return snapUpload(itemHelper, req.SourceAccountIDs, fmt.Sprintf("%s.tar.gz", rootDir))
				},
				nil,
			)
			req.InterruptStep = ""
		}
		if err := taskItem.Execute(); err != nil {
			_ = snapshotRepo.Update(req.ID, map[string]interface{}{"status": constant.StatusFailed, "message": err.Error(), "interrupt_step": taskItem.Task.CurrentStep})
			return
		}
		_ = snapshotRepo.Update(req.ID, map[string]interface{}{"status": constant.StatusSuccess, "interrupt_step": ""})
		_ = os.RemoveAll(rootDir)
	}()
	return nil
}

type snapHelper struct {
	SnapID      uint
	snapAgentDB *gorm.DB
	snapCoreDB  *gorm.DB
	Ctx         context.Context
	FileOp      files.FileOp
	Wg          *sync.WaitGroup
	Task        task.Task
}

func loadDbConn(snap *snapHelper, targetDir string, req dto.SnapshotCreate) error {
	snap.Task.Log("<######################## 1 / 8 ########################>")
	snap.Task.LogStart(i18n.GetMsgByKey("SnapDBInfo"))
	pathDB := path.Join(global.CONF.System.BaseDir, "1panel/db")

	err := snap.FileOp.CopyDir(pathDB, targetDir)
	snap.Task.LogWithStatus(i18n.GetWithName("SnapCopy", pathDB), err)
	if err != nil {
		return err
	}

	agentDb, err := newSnapDB(path.Join(targetDir, "db"), "agent.db")
	snap.Task.LogWithStatus(i18n.GetWithName("SnapNewDB", "agent"), err)
	if err != nil {
		return err
	}
	snap.snapAgentDB = agentDb
	coreDb, err := newSnapDB(path.Join(targetDir, "db"), "core.db")
	snap.Task.LogWithStatus(i18n.GetWithName("SnapNewDB", "core"), err)
	if err != nil {
		return err
	}
	snap.snapCoreDB = coreDb

	if !req.WithMonitorData {
		err = os.Remove(path.Join(targetDir, "db/monitor.db"))
		snap.Task.LogWithStatus(i18n.GetMsgByKey("SnapDeleteMonitor"), err)
		if err != nil {
			return err
		}
	}
	if !req.WithOperationLog {
		err = snap.snapCoreDB.Exec("DELETE FROM operation_logs").Error
		snap.Task.LogWithStatus(i18n.GetMsgByKey("SnapDeleteOperationLog"), err)
		if err != nil {
			return err
		}
	}
	if !req.WithLoginLog {
		err = snap.snapCoreDB.Exec("DELETE FROM login_logs").Error
		snap.Task.LogWithStatus(i18n.GetMsgByKey("SnapDeleteLoginLog"), err)
		if err != nil {
			return err
		}
	}

	_ = snap.snapAgentDB.Model(&model.Setting{}).Where("key = ?", "SystemIP").Updates(map[string]interface{}{"value": ""}).Error
	_ = snap.snapAgentDB.Where("id = ?", snap.SnapID).Delete(&model.Snapshot{}).Error

	return nil
}

func snapBaseData(snap snapHelper, targetDir string) error {
	snap.Task.Log("<######################## 2 / 8 ########################>")
	snap.Task.LogStart(i18n.GetMsgByKey("SnapBaseInfo"))

	err := common.CopyFile("/usr/local/bin/1panel", targetDir)
	snap.Task.LogWithStatus(i18n.GetWithName("SnapCopy", "/usr/local/bin/1panel"), err)
	if err != nil {
		return err
	}

	err = common.CopyFile("/usr/local/bin/1panel_agent", targetDir)
	snap.Task.LogWithStatus(i18n.GetWithName("SnapCopy", "/usr/local/bin/1panel_agent"), err)
	if err != nil {
		return err
	}

	err = common.CopyFile("/usr/local/bin/1pctl", targetDir)
	snap.Task.LogWithStatus(i18n.GetWithName("SnapCopy", "/usr/local/bin/1pctl"), err)
	if err != nil {
		return err
	}

	err = common.CopyFile("/etc/systemd/system/1panel.service", targetDir)
	snap.Task.LogWithStatus(i18n.GetWithName("SnapCopy", "/etc/systemd/system/1panel.service"), err)
	if err != nil {
		return err
	}

	err = common.CopyFile("/etc/systemd/system/1panel_agent.service", targetDir)
	snap.Task.LogWithStatus(i18n.GetWithName("SnapCopy", "/etc/systemd/system/1panel_agent.service"), err)
	if err != nil {
		return err
	}

	if snap.FileOp.Stat("/etc/docker/daemon.json") {
		err = common.CopyFile("/etc/docker/daemon.json", targetDir)
		snap.Task.LogWithStatus(i18n.GetWithName("SnapCopy", "/etc/docker/daemon.json"), err)
		if err != nil {
			return err
		}
	}

	remarkInfo, _ := json.MarshalIndent(SnapshotJson{
		BaseDir:       global.CONF.System.BaseDir,
		BackupDataDir: global.CONF.System.Backup,
	}, "", "\t")
	err = os.WriteFile(path.Join(targetDir, "snapshot.json"), remarkInfo, 0640)
	snap.Task.LogWithStatus(i18n.GetWithName("SnapCopy", path.Join(targetDir, "snapshot.json")), err)
	if err != nil {
		return err
	}

	return nil
}

func snapAppImage(snap snapHelper, req dto.SnapshotCreate, targetDir string) error {
	snap.Task.Log("<######################## 3 / 8 ########################>")
	snap.Task.LogStart(i18n.GetMsgByKey("SnapInstallApp"))

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
		snap.Task.Logf("docker save %s | gzip -c > %s", strings.Join(imageList, " "), path.Join(targetDir, "images.tar.gz"))
		std, err := cmd.Execf("docker save %s | gzip -c > %s", strings.Join(imageList, " "), path.Join(targetDir, "images.tar.gz"))
		snap.Task.LogWithStatus(i18n.GetMsgByKey("SnapDockerSave"), errors.New(std))
		if err != nil {
			return errors.New(std)
		}
	}
	return nil
}

func snapBackupData(snap snapHelper, req dto.SnapshotCreate, targetDir string) error {
	snap.Task.Log("<######################## 4 / 8 ########################>")
	snap.Task.LogStart(i18n.GetMsgByKey("SnapLocalBackup"))

	excludes := loadBackupExcludes(snap, req.BackupData)
	for _, item := range req.AppData {
		for _, itemApp := range item.Children {
			if itemApp.Label == "appBackup" {
				excludes = append(excludes, loadAppBackupExcludes([]dto.DataTree{itemApp})...)
			}
		}
	}
	err := snap.FileOp.TarGzCompressPro(false, global.CONF.System.Backup, path.Join(targetDir, "1panel_backup.tar.gz"), "", strings.Join(excludes, ";"))
	snap.Task.LogWithStatus(i18n.GetMsgByKey("SnapCompressBackup"), err)

	return err
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
					snap.Task.LogWithStatus("delete snapshot from database", err)
				}
			} else {
				fmt.Println(strings.TrimPrefix(path.Dir(item.Path), global.CONF.System.Backup+"/"), path.Base(item.Path))
				if err := snap.snapAgentDB.Debug().Where("file_dir = ? AND file_name = ?", strings.TrimPrefix(path.Dir(item.Path), global.CONF.System.Backup+"/"), path.Base(item.Path)).Delete(&model.BackupRecord{}).Error; err != nil {
					snap.Task.LogWithStatus("delete backup file from database", err)
				}
			}
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

func snapPanelData(snap snapHelper, req dto.SnapshotCreate, targetDir string) error {
	snap.Task.Log("<######################## 5 / 8 ########################>")
	snap.Task.LogStart(i18n.GetMsgByKey("SnapPanelData"))

	excludes := loadPanelExcludes(req.PanelData)
	for _, item := range req.AppData {
		for _, itemApp := range item.Children {
			if itemApp.Label == "appData" {
				excludes = append(excludes, loadPanelExcludes([]dto.DataTree{itemApp})...)
			}
		}
	}
	excludes = append(excludes, "./tmp")
	excludes = append(excludes, "./cache")
	excludes = append(excludes, "./uploads")
	excludes = append(excludes, "./db")
	excludes = append(excludes, "./resource")
	if !req.WithSystemLog {
		excludes = append(excludes, "./log/1Panel*")
	}
	if !req.WithTaskLog {
		excludes = append(excludes, "./log/App")
		excludes = append(excludes, "./log/Snapshot")
		excludes = append(excludes, "./log/AppStore")
		excludes = append(excludes, "./log/Website")
	}

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
	err := snap.FileOp.TarGzCompressPro(false, rootDir, path.Join(targetDir, "1panel_data.tar.gz"), "", strings.Join(excludes, ";"))
	snap.Task.LogWithStatus(i18n.GetMsgByKey("SnapCompressPanel"), err)

	return err
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

func snapCompress(snap snapHelper, rootDir string, secret string) error {
	snap.Task.Log("<######################## 7 / 8 ########################>")
	snap.Task.LogStart(i18n.GetMsgByKey("SnapCompress"))

	tmpDir := path.Join(global.CONF.System.TmpDir, "system")
	fileName := fmt.Sprintf("%s.tar.gz", path.Base(rootDir))
	err := snap.FileOp.TarGzCompressPro(true, rootDir, path.Join(tmpDir, fileName), secret, "")
	snap.Task.LogWithStatus(i18n.GetMsgByKey("SnapCompressFile"), err)
	if err != nil {
		return err
	}

	stat, err := os.Stat(path.Join(tmpDir, fileName))
	snap.Task.LogWithStatus(i18n.GetMsgByKey("SnapCheckCompress"), err)
	if err != nil {
		return err
	}

	size := common.LoadSizeUnit2F(float64(stat.Size()))
	snap.Task.Logf(i18n.GetWithName("SnapCompressSize", size))
	_ = os.RemoveAll(rootDir)
	return nil
}

func snapUpload(snap snapHelper, accounts string, file string) error {
	snap.Task.Log("<######################## 8 / 8 ########################>")
	snap.Task.LogStart(i18n.GetMsgByKey("SnapUpload"))

	source := path.Join(global.CONF.System.TmpDir, "system", path.Base(file))
	accountMap, err := NewBackupClientMap(strings.Split(accounts, ","))
	snap.Task.LogWithStatus(i18n.GetMsgByKey("SnapLoadBackup"), err)
	if err != nil {
		return err
	}

	targetAccounts := strings.Split(accounts, ",")
	for _, item := range targetAccounts {
		snap.Task.LogStart(i18n.GetWithName("SnapUploadTo", fmt.Sprintf("[%s] %s", accountMap[item].name, path.Join(accountMap[item].backupPath, "system_snapshot", path.Base(file)))))
		_, err := accountMap[item].client.Upload(source, path.Join(accountMap[item].backupPath, "system_snapshot", path.Base(file)))
		snap.Task.LogWithStatus(i18n.GetWithName("SnapUploadRes", accountMap[item].name), err)
		if err != nil {
			return err
		}
	}
	_ = os.Remove(source)
	return nil
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
	return db, nil
}

func closeDatabase(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		return
	}
	_ = sqlDB.Close()
}

func rebuildAllAppInstall() error {
	global.LOG.Debug("start to rebuild all app")
	appInstalls, err := appInstallRepo.ListBy()
	if err != nil {
		global.LOG.Errorf("get all app installed for rebuild failed, err: %v", err)
		return err
	}
	var wg sync.WaitGroup
	for i := 0; i < len(appInstalls); i++ {
		wg.Add(1)
		appInstalls[i].Status = constant.Rebuilding
		_ = appInstallRepo.Save(context.Background(), &appInstalls[i])
		go func(app model.AppInstall) {
			defer wg.Done()
			dockerComposePath := app.GetComposePath()
			out, err := compose.Down(dockerComposePath)
			if err != nil {
				_ = handleErr(app, err, out)
				return
			}
			out, err = compose.Up(dockerComposePath)
			if err != nil {
				_ = handleErr(app, err, out)
				return
			}
			app.Status = constant.Running
			_ = appInstallRepo.Save(context.Background(), &app)
		}(appInstalls[i])
	}
	wg.Wait()
	return nil
}
