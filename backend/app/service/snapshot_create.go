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
	if err := common.CopyFile("/usr/local/bin/1panel", path.Join(targetDir, "1panel")); err != nil {
		status = err.Error()
	}

	if err := common.CopyFile("/usr/local/bin/1pctl", targetDir); err != nil {
		status = err.Error()
	}
	_, _ = cmd.Execf("cp -r /usr/local/bin/lang %s", targetDir)

	if err := common.CopyFile("/etc/systemd/system/1panel.service", targetDir); err != nil {
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
	if err := common.CopyFile("/etc/docker/daemon.json", targetDir); err != nil {
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
	if err := handleSnapTar(localDir, targetDir, "1panel_backup.tar.gz", "./system;./system_snapshot;", ""); err != nil {
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
	ignoreVal, _ := settingRepo.Get(settingRepo.WithByKey("SnapshotIgnore"))
	rules := strings.Split(ignoreVal.Value, ",")
	for _, ignore := range rules {
		if len(ignore) == 0 || cmd.CheckIllegal(ignore) {
			continue
		}
		exclusionRules += ("." + strings.ReplaceAll(ignore, dataDir, "") + ";")
	}
	_ = snapshotRepo.Update(snap.SnapID, map[string]interface{}{"status": "OnSaveData"})
	sysIP, _ := settingRepo.Get(settingRepo.WithByKey("SystemIP"))
	_ = settingRepo.Update("SystemIP", "")
	checkPointOfWal()
	if err := handleSnapTar(dataDir, targetDir, "1panel_data.tar.gz", exclusionRules, ""); err != nil {
		status = err.Error()
	}
	_ = snapshotRepo.Update(snap.SnapID, map[string]interface{}{"status": constant.StatusWaiting})

	snap.Status.PanelData = status
	_ = snapshotRepo.UpdateStatus(snap.Status.ID, map[string]interface{}{"panel_data": status})
	_ = settingRepo.Update("SystemIP", sysIP.Value)
}

func snapCompress(snap snapHelper, rootDir string, secret string) {
	_ = snapshotRepo.UpdateStatus(snap.Status.ID, map[string]interface{}{"compress": constant.StatusRunning})
	tmpDir := path.Join(global.CONF.System.TmpDir, "system")
	fileName := fmt.Sprintf("%s.tar.gz", path.Base(rootDir))
	if err := handleSnapTar(rootDir, tmpDir, fileName, "", secret); err != nil {
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

	global.LOG.Debugf("remove snapshot file %s", source)
	_ = os.Remove(source)
}

func handleSnapTar(sourceDir, targetDir, name, exclusionRules string, secret string) error {
	if _, err := os.Stat(targetDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(targetDir, os.ModePerm); err != nil {
			return err
		}
	}

	exMap := make(map[string]struct{})
	exStr := ""
	excludes := strings.Split(exclusionRules, ";")
	for _, exclude := range excludes {
		if len(exclude) == 0 {
			continue
		}
		if _, ok := exMap[exclude]; ok {
			continue
		}
		exStr += " --exclude "
		exStr += exclude
		exMap[exclude] = struct{}{}
	}
	path := ""
	if strings.Contains(sourceDir, "/") {
		itemDir := strings.ReplaceAll(sourceDir[strings.LastIndex(sourceDir, "/"):], "/", "")
		aheadDir := sourceDir[:strings.LastIndex(sourceDir, "/")]
		if len(aheadDir) == 0 {
			aheadDir = "/"
		}
		path += fmt.Sprintf("-C %s %s", aheadDir, itemDir)
	} else {
		path = sourceDir
	}
	commands := ""
	if len(secret) != 0 {
		extraCmd := "| openssl enc -aes-256-cbc -salt -k '" + secret + "' -out"
		commands = fmt.Sprintf("tar --warning=no-file-changed --ignore-failed-read --exclude-from=<(find %s -type s -print) -zcf %s %s %s %s", sourceDir, " -"+exStr, path, extraCmd, targetDir+"/"+name)
		global.LOG.Debug(strings.ReplaceAll(commands, fmt.Sprintf(" %s ", secret), "******"))
	} else {
		commands = fmt.Sprintf("tar --warning=no-file-changed --ignore-failed-read --exclude-from=<(find %s -type s -printf '%s' | sed 's|^|./|') -zcf %s %s -C %s .", sourceDir, "%P\n", targetDir+"/"+name, exStr, sourceDir)
		global.LOG.Debug(commands)
	}
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
