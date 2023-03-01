package service

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/compose"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func (u *BackupService) AppBackup(req dto.CommonBackup) error {
	localDir, err := loadLocalDir()
	if err != nil {
		return err
	}
	app, err := appRepo.GetFirst(appRepo.WithKey(req.Name))
	if err != nil {
		return err
	}
	install, err := appInstallRepo.GetFirst(commonRepo.WithByName(req.DetailName), appInstallRepo.WithAppId(app.ID))
	if err != nil {
		return err
	}
	timeNow := time.Now().Format("20060102150405")
	backupDir := fmt.Sprintf("%s/app/%s/%s", localDir, req.Name, req.DetailName)

	fileName := fmt.Sprintf("%s_%s.tar.gz", req.DetailName, timeNow)
	if err := handleAppBackup(&install, backupDir, fileName); err != nil {
		return err
	}

	record := &model.BackupRecord{
		Type:       "app",
		Name:       req.Name,
		DetailName: req.DetailName,
		Source:     "LOCAL",
		BackupType: "LOCAL",
		FileDir:    backupDir,
		FileName:   fileName,
	}

	if err := backupRepo.CreateRecord(record); err != nil {
		global.LOG.Errorf("save backup record failed, err: %v", err)
		return err
	}
	return nil
}

func (u *BackupService) AppRecover(req dto.CommonRecover) error {
	app, err := appRepo.GetFirst(appRepo.WithKey(req.Name))
	if err != nil {
		return err
	}
	install, err := appInstallRepo.GetFirst(commonRepo.WithByName(req.DetailName), appInstallRepo.WithAppId(app.ID))
	if err != nil {
		return err
	}

	fileOp := files.NewFileOp()
	if !fileOp.Stat(req.File) {
		return errors.New(fmt.Sprintf("%s file is not exist", req.File))
	}
	if _, err := compose.Down(install.GetComposePath()); err != nil {
		return err
	}
	if err := handleAppRecover(&install, req.File, false); err != nil {
		return err
	}
	return nil
}

type AppInfo struct {
	AppDetailId uint   `json:"appDetailId"`
	Param       string `json:"param"`
	Version     string `json:"version"`
}

func handleAppBackup(install *model.AppInstall, backupDir, fileName string) error {
	fileOp := files.NewFileOp()
	tmpDir := fmt.Sprintf("%s/%s", backupDir, strings.ReplaceAll(fileName, ".tar.gz", ""))
	if !fileOp.Stat(tmpDir) {
		if err := os.MkdirAll(tmpDir, os.ModePerm); err != nil {
			return fmt.Errorf("mkdir %s failed, err: %v", backupDir, err)
		}
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	var appInfo AppInfo
	appInfo.Param = install.Param
	appInfo.AppDetailId = install.AppDetailId
	appInfo.Version = install.Version
	remarkInfo, _ := json.Marshal(appInfo)
	remarkInfoPath := fmt.Sprintf("%s/app.json", tmpDir)
	if err := fileOp.SaveFile(remarkInfoPath, string(remarkInfo), fs.ModePerm); err != nil {
		return err
	}

	appPath := fmt.Sprintf("%s/%s/%s", constant.AppInstallDir, install.App.Key, install.Name)
	if err := fileOp.Compress([]string{appPath}, tmpDir, "app.tar.gz", files.TarGz); err != nil {
		return err
	}
	if err := fileOp.Compress([]string{tmpDir}, backupDir, fileName, files.TarGz); err != nil {
		return err
	}
	return nil
}

func handleAppRecover(install *model.AppInstall, recoverFile string, isRollback bool) error {
	isOk := false
	fileOp := files.NewFileOp()
	if err := fileOp.Decompress(recoverFile, path.Dir(recoverFile), files.TarGz); err != nil {
		return err
	}
	tmpPath := strings.ReplaceAll(recoverFile, ".tar.gz", "")
	defer func() {
		_ = os.RemoveAll(strings.ReplaceAll(recoverFile, ".tar.gz", ""))
	}()

	if !fileOp.Stat(tmpPath+"/app.json") || !fileOp.Stat(tmpPath+"/app.tar.gz") {
		return errors.New("the wrong recovery package does not have app.json or app.tar.gz files")
	}
	if !isRollback {
		rollbackFile := fmt.Sprintf("%s/original/app/%s_%s.tar.gz", global.CONF.System.BaseDir, install.Name, time.Now().Format("20060102150405"))
		if err := handleAppBackup(install, path.Dir(rollbackFile), path.Base(rollbackFile)); err != nil {
			global.LOG.Errorf("backup app %s for rollback before recover failed, err: %v", install.Name, err)
		}
		defer func() {
			if !isOk {
				if err := handleAppRecover(install, rollbackFile, true); err != nil {
					global.LOG.Errorf("rollback app %s from %s failed, err: %v", install.Name, rollbackFile, err)
				}
				global.LOG.Infof("rollback app %s from %s successful", install.Name, rollbackFile)
				_ = os.RemoveAll(rollbackFile)
			} else {
				_ = os.RemoveAll(rollbackFile)
			}
		}()
	}

	appjson, err := os.ReadFile(tmpPath + "/" + "app.json")
	if err != nil {
		return err
	}
	var appInfo AppInfo
	_ = json.Unmarshal(appjson, &appInfo)

	if err := fileOp.Decompress(tmpPath+"/app.tar.gz", fmt.Sprintf("%s/%s", constant.AppInstallDir, install.App.Key), files.TarGz); err != nil {
		return err
	}
	composeContent, err := os.ReadFile(install.GetComposePath())
	if err != nil {
		return err
	}
	install.DockerCompose = string(composeContent)
	envContent, err := os.ReadFile(fmt.Sprintf("%s/%s/%s/.env", constant.AppInstallDir, install.App.Key, install.Name))
	if err != nil {
		return err
	}
	install.Env = string(envContent)
	envMaps, err := godotenv.Unmarshal(string(envContent))
	if err != nil {
		return err
	}
	install.HttpPort = 0
	httpPort, ok := envMaps["PANEL_APP_PORT_HTTP"]
	if ok {
		httpPortN, _ := strconv.Atoi(httpPort)
		install.HttpPort = httpPortN
	}
	install.HttpsPort = 0
	httpsPort, ok := envMaps["PANEL_APP_PORT_HTTPS"]
	if ok {
		httpsPortN, _ := strconv.Atoi(httpsPort)
		install.HttpsPort = httpsPortN
	}

	composeMap := make(map[string]interface{})
	if err := yaml.Unmarshal(composeContent, &composeMap); err != nil {
		return err
	}
	servicesMap := composeMap["services"].(map[string]interface{})
	for k, v := range servicesMap {
		install.ServiceName = k
		value := v.(map[string]interface{})
		install.ContainerName = value["container_name"].(string)
	}

	install.Param = appInfo.Param
	if out, err := compose.Up(install.GetComposePath()); err != nil {
		install.Message = err.Error()
		if len(out) != 0 {
			install.Message = out
		}
		return errors.New(out)
	}
	install.AppDetailId = appInfo.AppDetailId
	install.Version = appInfo.Version
	install.Status = constant.Running
	if err := appInstallRepo.Save(install); err != nil {
		return err
	}
	isOk = true
	return nil
}
