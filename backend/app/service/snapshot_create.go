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

	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
)

type snapHelper struct {
	SnapID uint
	Ctx    context.Context
	FileOp files.FileOp
	Wg     *sync.WaitGroup
}

func snapJson(snap snapHelper, statusID uint, snapJson SnapshotJson, targetDir string) {
	defer snap.Wg.Done()
	_ = snapshotRepo.UpdateStatus(statusID, map[string]interface{}{"panel_info": constant.Running})
	status := constant.StatusDone
	remarkInfo, _ := json.MarshalIndent(snapJson, "", "\t")
	if err := os.WriteFile(fmt.Sprintf("%s/snapshot.json", targetDir), remarkInfo, 0640); err != nil {
		status = err.Error()
	}
	_ = snapshotRepo.UpdateStatus(statusID, map[string]interface{}{"panel_info": status})
}

func snapPanel(snap snapHelper, statusID uint, targetDir string) {
	defer snap.Wg.Done()
	_ = snapshotRepo.UpdateStatus(statusID, map[string]interface{}{"panel": constant.Running})
	status := constant.StatusDone
	if err := cpBinary("/usr/local/bin/1panel", path.Join(targetDir, "1panel")); err != nil {
		status = err.Error()
	}
	_ = snapshotRepo.UpdateStatus(statusID, map[string]interface{}{"panel": status})
}

func snapPanelCtl(snap snapHelper, statusID uint, targetDir string) {
	defer snap.Wg.Done()
	_ = snapshotRepo.UpdateStatus(statusID, map[string]interface{}{"panel_ctl": constant.Running})
	status := constant.StatusDone
	if err := cpBinary("/usr/local/bin/1pctl", path.Join(targetDir, "1pctl")); err != nil {
		status = err.Error()
	}
	_ = snapshotRepo.UpdateStatus(statusID, map[string]interface{}{"panel_ctl": status})
}

func snapPanelService(snap snapHelper, statusID uint, targetDir string) {
	defer snap.Wg.Done()
	_ = snapshotRepo.UpdateStatus(statusID, map[string]interface{}{"panel_service": constant.Running})
	status := constant.StatusDone
	if err := cpBinary("/etc/systemd/system/1panel.service", path.Join(targetDir, "1panel.service")); err != nil {
		status = err.Error()
	}
	_ = snapshotRepo.UpdateStatus(statusID, map[string]interface{}{"panel_service": status})
}

func snapDaemonJson(snap snapHelper, statusID uint, targetDir string) {
	defer snap.Wg.Done()
	if !snap.FileOp.Stat("/etc/docker/daemon.json") {
		return
	}
	_ = snapshotRepo.UpdateStatus(statusID, map[string]interface{}{"daemon_json": constant.Running})
	status := constant.StatusDone
	if err := cpBinary("/etc/docker/daemon.json", path.Join(targetDir, "daemon.json")); err != nil {
		status = err.Error()
	}
	_ = snapshotRepo.UpdateStatus(statusID, map[string]interface{}{"daemon_json": status})
}

func snapAppData(snap snapHelper, statusID uint, targetDir string) {
	defer snap.Wg.Done()
	_ = snapshotRepo.UpdateStatus(statusID, map[string]interface{}{"app_data": constant.Running})
	appInstalls, err := appInstallRepo.ListBy()
	if err != nil {
		_ = snapshotRepo.UpdateStatus(statusID, map[string]interface{}{"app_data": err.Error()})
		return
	}
	runtimes, err := runtimeRepo.List()
	if err != nil {
		_ = snapshotRepo.UpdateStatus(statusID, map[string]interface{}{"app_data": err.Error()})
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
	for _, rumtime := range runtimes {
		for _, existImage := range existImages {
			if rumtime.Image == existImage && !duplicateMap[rumtime.Image] {
				imageSaveList = append(imageSaveList, rumtime.Image)
				duplicateMap[rumtime.Image] = true
			}
		}
	}

	global.LOG.Debugf("docker save %s | gzip -c > %s", strings.Join(imageSaveList, " "), path.Join(targetDir, "docker_image.tar"))
	std, err := cmd.Execf("docker save %s | gzip -c > %s", strings.Join(imageSaveList, " "), path.Join(targetDir, "docker_image.tar"))
	if err != nil {
		_ = snapshotRepo.UpdateStatus(statusID, map[string]interface{}{"app_data": std})
		return
	}
	_ = snapshotRepo.UpdateStatus(statusID, map[string]interface{}{"app_data": constant.StatusDone})
}

func snapBackup(snap snapHelper, statusID uint, localDir, targetDir string) {
	defer snap.Wg.Done()
	_ = snapshotRepo.UpdateStatus(statusID, map[string]interface{}{"backup_data": constant.Running})
	status := constant.StatusDone
	if err := handleSnapTar(localDir, targetDir, "1panel_backup.tar.gz", "./system;"); err != nil {
		status = err.Error()
	}
	_ = snapshotRepo.UpdateStatus(statusID, map[string]interface{}{"backup_data": status})
}

func snapPanelData(snap snapHelper, statusID uint, localDir, targetDir string) {
	defer snap.Wg.Done()
	_ = snapshotRepo.UpdateStatus(statusID, map[string]interface{}{"panel_data": constant.Running})
	status := constant.StatusDone
	dataDir := path.Join(global.CONF.System.BaseDir, "1panel")
	exclusionRules := "./tmp;./log;./cache;./db/1Panel.db-*;"
	if strings.Contains(localDir, dataDir) {
		exclusionRules += ("." + strings.ReplaceAll(localDir, dataDir, "") + ";")
	}

	_ = snapshotRepo.Update(snap.SnapID, map[string]interface{}{"status": "OnSaveData"})
	if err := handleSnapTar(dataDir, targetDir, "1panel_data.tar.gz", exclusionRules); err != nil {
		status = err.Error()
	}
	_ = snapshotRepo.Update(snap.SnapID, map[string]interface{}{"status": constant.StatusWaiting})
	_ = snapshotRepo.UpdateStatus(statusID, map[string]interface{}{"panel_data": status})
}

func snapCompress(snap snapHelper, statusID uint, rootDir string) {
	defer func() {
		global.LOG.Debugf("remove snapshot file %s", rootDir)
		_ = os.RemoveAll(rootDir)
	}()
	_ = snapshotRepo.UpdateStatus(statusID, map[string]interface{}{"compress": constant.StatusRunning})
	tmpDir := path.Join(global.CONF.System.TmpDir, "system")
	fileName := fmt.Sprintf("%s.tar.gz", path.Base(rootDir))
	if err := snap.FileOp.Compress([]string{rootDir}, tmpDir, fileName, files.TarGz); err != nil {
		_ = snapshotRepo.UpdateStatus(statusID, map[string]interface{}{"compress": err.Error()})
		return
	}
	_ = snapshotRepo.UpdateStatus(statusID, map[string]interface{}{"compress": constant.StatusDone})
}

func snapUpload(account string, statusID uint, file string) {
	source := path.Join(global.CONF.System.TmpDir, "system", path.Base(file))
	defer func() {
		global.LOG.Debugf("remove snapshot file %s", source)
		_ = os.Remove(source)
	}()

	_ = snapshotRepo.UpdateStatus(statusID, map[string]interface{}{"upload": constant.StatusUploading})
	backup, err := backupRepo.Get(commonRepo.WithByType(account))
	if err != nil {
		_ = snapshotRepo.UpdateStatus(statusID, map[string]interface{}{"upload": err.Error()})
		return
	}
	client, err := NewIBackupService().NewClient(&backup)
	if err != nil {
		_ = snapshotRepo.UpdateStatus(statusID, map[string]interface{}{"upload": err.Error()})
		return
	}
	target := path.Join(backup.BackupPath, "system_snapshot", path.Base(file))
	if _, err := client.Upload(source, target); err != nil {
		_ = snapshotRepo.UpdateStatus(statusID, map[string]interface{}{"upload": err.Error()})
		return
	}
	_ = snapshotRepo.UpdateStatus(statusID, map[string]interface{}{"upload": constant.StatusDone})
}

func handleSnapTar(sourceDir, targetDir, name, exclusionRules string) error {
	if _, err := os.Stat(targetDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(targetDir, os.ModePerm); err != nil {
			return err
		}
	}

	exStr := ""
	excludes := strings.Split(exclusionRules, ";")
	for _, exclude := range excludes {
		if len(exclude) == 0 {
			continue
		}
		exStr += " --exclude "
		exStr += exclude
	}

	commands := fmt.Sprintf("tar --warning=no-file-changed -zcf %s %s -C %s .", targetDir+"/"+name, exStr, sourceDir)
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
