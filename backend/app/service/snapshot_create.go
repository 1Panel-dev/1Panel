package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/common"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
)

type snapHelper struct {
	SnapID uint
	Status *model.SnapshotStatus
	Ctx    context.Context
	FileOp files.FileOp
	Wg     *sync.WaitGroup
}

func snapJson(snap snapHelper, snapJson SnapshotJson, targetDir string) {
	defer snap.Wg.Done()
	_ = snapshotRepo.UpdateStatus(snap.Status.ID, map[string]interface{}{"panel_info": constant.Running})
	status := constant.StatusDone
	remarkInfo, _ := json.MarshalIndent(snapJson, "", "\t")
	if err := os.WriteFile(fmt.Sprintf("%s/snapshot.json", targetDir), remarkInfo, 0640); err != nil {
		status = err.Error()
	}
	snap.Status.PanelInfo = status
	_ = snapshotRepo.UpdateStatus(snap.Status.ID, map[string]interface{}{"panel_info": status})
}

func snapPanel(snap snapHelper, targetDir string) {
	defer snap.Wg.Done()
	_ = snapshotRepo.UpdateStatus(snap.Status.ID, map[string]interface{}{"panel": constant.Running})
	status := constant.StatusDone
	if err := cpBinary([]string{"/usr/local/bin/1panel", "/usr/local/bin/1pctl", "/etc/systemd/system/1panel.service"}, targetDir); err != nil {
		status = err.Error()
	}
	snap.Status.Panel = status
	_ = snapshotRepo.UpdateStatus(snap.Status.ID, map[string]interface{}{"panel": status})
}

func snapDaemonJson(snap snapHelper, targetDir string) {
	defer snap.Wg.Done()
	status := constant.StatusDone
	if !snap.FileOp.Stat("/etc/docker/daemon.json") {
		snap.Status.DaemonJson = status
		_ = snapshotRepo.UpdateStatus(snap.Status.ID, map[string]interface{}{"daemon_json": status})
		return
	}
	_ = snapshotRepo.UpdateStatus(snap.Status.ID, map[string]interface{}{"daemon_json": constant.Running})
	if err := cpBinary([]string{"/etc/docker/daemon.json"}, path.Join(targetDir, "daemon.json")); err != nil {
		status = err.Error()
	}
	snap.Status.DaemonJson = status
	_ = snapshotRepo.UpdateStatus(snap.Status.ID, map[string]interface{}{"daemon_json": status})
}

func snapAppData(snap snapHelper, targetDir string) {
	defer snap.Wg.Done()
	_ = snapshotRepo.UpdateStatus(snap.Status.ID, map[string]interface{}{"app_data": constant.Running})
	appInstalls, err := appInstallRepo.ListBy()
	if err != nil {
		snap.Status.AppData = err.Error()
		_ = snapshotRepo.UpdateStatus(snap.Status.ID, map[string]interface{}{"app_data": err.Error()})
		return
	}
	runtimes, err := runtimeRepo.List()
	if err != nil {
		snap.Status.AppData = err.Error()
		_ = snapshotRepo.UpdateStatus(snap.Status.ID, map[string]interface{}{"app_data": err.Error()})
		return
	}
	imageRegex := regexp.MustCompile(`image:\s*(.*)`)
	var imageSaveList []string
	existStr, _ := cmd.Exec("docker images | awk '{print $1\":\"$2}' | grep -v REPOSITORY:TAG")
	existImages := strings.Split(existStr, "\n")
	duplicateMap := make(map[string]bool)
	for _, app := range appInstalls {
		matches := imageRegex.FindAllStringSubmatch(app.DockerCompose, -1)
		for _, match := range matches {
			for _, existImage := range existImages {
				if match[1] == existImage && !duplicateMap[match[1]] {
					imageSaveList = append(imageSaveList, match[1])
					duplicateMap[match[1]] = true
				}
			}
		}
	}
	for _, runtime := range runtimes {
		for _, existImage := range existImages {
			if runtime.Image == existImage && !duplicateMap[runtime.Image] {
				imageSaveList = append(imageSaveList, runtime.Image)
				duplicateMap[runtime.Image] = true
			}
		}
	}

	if len(imageSaveList) != 0 {
		global.LOG.Debugf("docker save %s | gzip -c > %s", strings.Join(imageSaveList, " "), path.Join(targetDir, "docker_image.tar"))
		std, err := cmd.Execf("docker save %s | gzip -c > %s", strings.Join(imageSaveList, " "), path.Join(targetDir, "docker_image.tar"))
		if err != nil {
			snap.Status.AppData = err.Error()
			_ = snapshotRepo.UpdateStatus(snap.Status.ID, map[string]interface{}{"app_data": std})
			return
		}
	}
	snap.Status.AppData = constant.StatusDone
	_ = snapshotRepo.UpdateStatus(snap.Status.ID, map[string]interface{}{"app_data": constant.StatusDone})
}

func snapBackup(snap snapHelper, localDir, targetDir string) {
	defer snap.Wg.Done()
	_ = snapshotRepo.UpdateStatus(snap.Status.ID, map[string]interface{}{"backup_data": constant.Running})
	status := constant.StatusDone
	if err := handleSnapTar(localDir, targetDir, "1panel_backup.tar.gz", "./system;./system_snapshot;"); err != nil {
		status = err.Error()
	}
	snap.Status.BackupData = status
	_ = snapshotRepo.UpdateStatus(snap.Status.ID, map[string]interface{}{"backup_data": status})
}

func snapPanelData(snap snapHelper, localDir, targetDir string) {
	_ = snapshotRepo.UpdateStatus(snap.Status.ID, map[string]interface{}{"panel_data": constant.Running})
	status := constant.StatusDone
	dataDir := path.Join(global.CONF.System.BaseDir, "1panel")
	exclusionRules := "./tmp;./log;./cache;./db/1Panel.db-*;"
	if strings.Contains(localDir, dataDir) {
		exclusionRules += ("." + strings.ReplaceAll(localDir, dataDir, "") + ";")
	}

	_ = snapshotRepo.Update(snap.SnapID, map[string]interface{}{"status": "OnSaveData"})
	checkPointOfWal()
	if err := handleSnapTar(dataDir, targetDir, "1panel_data.tar.gz", exclusionRules); err != nil {
		status = err.Error()
	}
	_ = snapshotRepo.Update(snap.SnapID, map[string]interface{}{"status": constant.StatusWaiting})

	snap.Status.PanelData = status
	_ = snapshotRepo.UpdateStatus(snap.Status.ID, map[string]interface{}{"panel_data": status})
}

func snapCompress(snap snapHelper, rootDir string) {
	defer func() {
		global.LOG.Debugf("remove snapshot file %s", rootDir)
		_ = os.RemoveAll(rootDir)
	}()
	_ = snapshotRepo.UpdateStatus(snap.Status.ID, map[string]interface{}{"compress": constant.StatusRunning})
	tmpDir := path.Join(global.CONF.System.TmpDir, "system")
	fileName := fmt.Sprintf("%s.tar.gz", path.Base(rootDir))
	if err := snap.FileOp.Compress([]string{rootDir}, tmpDir, fileName, files.TarGz); err != nil {
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
}

func snapUpload(snap snapHelper, accounts string, file string) {
	source := path.Join(global.CONF.System.TmpDir, "system", path.Base(file))
	defer func() {
		global.LOG.Debugf("remove snapshot file %s", source)
		_ = os.Remove(source)
	}()

	_ = snapshotRepo.UpdateStatus(snap.Status.ID, map[string]interface{}{"upload": constant.StatusUploading})
	accountMap, err := loadClientMap(accounts)
	if err != nil {
		snap.Status.Upload = err.Error()
		_ = snapshotRepo.UpdateStatus(snap.Status.ID, map[string]interface{}{"upload": err.Error()})
		return
	}
	targetAccounts := strings.Split(accounts, ",")
	for _, item := range targetAccounts {
		global.LOG.Debugf("start upload snapshot to %s, path: %s", item, path.Join(accountMap[item].backupPath, "system_snapshot", path.Base(file)))
		if _, err := accountMap[item].client.Upload(source, path.Join(accountMap[item].backupPath, "system_snapshot", path.Base(file))); err != nil {
			global.LOG.Debugf("upload to %s failed, err: %v", item, err)
			snap.Status.Upload = err.Error()
			_ = snapshotRepo.UpdateStatus(snap.Status.ID, map[string]interface{}{"upload": err.Error()})
			return
		}
		global.LOG.Debugf("upload to %s successful", item)
	}
	snap.Status.Upload = constant.StatusDone
	_ = snapshotRepo.UpdateStatus(snap.Status.ID, map[string]interface{}{"upload": constant.StatusDone})
}

func handleSnapTar(sourceDir, targetDir, name, exclusionRules string) error {
	if _, err := os.Stat(targetDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(targetDir, os.ModePerm); err != nil {
			return err
		}
	}

	exStr := ""
	excludes := strings.Split(exclusionRules, ";")
	excludes = append(excludes, "*.sock")
	for _, exclude := range excludes {
		if len(exclude) == 0 {
			continue
		}
		exStr += " --exclude "
		exStr += exclude
	}

	commands := fmt.Sprintf("tar --warning=no-file-changed --ignore-failed-read -zcf %s %s -C %s .", targetDir+"/"+name, exStr, sourceDir)
	global.LOG.Debug(commands)
	stdout, err := cmd.ExecWithTimeOut(commands, 30*time.Minute)
	if err != nil {
		if len(stdout) != 0 {
			global.LOG.Errorf("do handle tar failed, stdout: %s, err: %v", stdout, err)
			return fmt.Errorf("do handle tar failed, stdout: %s, err: %v", stdout, err)
		}
	}
	return nil
}

func checkPointOfWal() {
	if err := global.DB.Exec("PRAGMA wal_checkpoint(TRUNCATE);").Error; err != nil {
		global.LOG.Errorf("handle check point failed, err: %v", err)
	}
}
