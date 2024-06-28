package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
)

func handleRuntimeBackup(runtime *model.Runtime, backupDir, fileName string, excludes string, secret string) error {
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

	remarkInfo, _ := json.Marshal(runtime)
	remarkInfoPath := fmt.Sprintf("%s/runtime.json", tmpDir)
	if err := fileOp.SaveFile(remarkInfoPath, string(remarkInfo), fs.ModePerm); err != nil {
		return err
	}

	appPath := runtime.GetPath()
	if err := handleTar(appPath, tmpDir, "runtime.tar.gz", excludes, secret); err != nil {
		return err
	}
	if err := handleTar(tmpDir, backupDir, fileName, "", secret); err != nil {
		return err
	}
	return nil
}

func handleRuntimeRecover(runtime *model.Runtime, recoverFile string, isRollback bool, secret string) error {
	isOk := false
	fileOp := files.NewFileOp()
	if err := handleUnTar(recoverFile, path.Dir(recoverFile), secret); err != nil {
		return err
	}
	tmpPath := strings.ReplaceAll(recoverFile, ".tar.gz", "")
	defer func() {
		go startRuntime(runtime)
		_ = os.RemoveAll(strings.ReplaceAll(recoverFile, ".tar.gz", ""))
	}()

	if !fileOp.Stat(tmpPath+"/runtime.json") || !fileOp.Stat(tmpPath+"/runtime.tar.gz") {
		return errors.New("the wrong recovery package does not have runtime.json or runtime.tar.gz files")
	}
	var oldRuntime model.Runtime
	runtimeJson, err := os.ReadFile(tmpPath + "/runtime.json")
	if err != nil {
		return err
	}
	if err := json.Unmarshal(runtimeJson, &oldRuntime); err != nil {
		return fmt.Errorf("unmarshal runtime.json failed, err: %v", err)
	}
	if oldRuntime.Type != runtime.Type || oldRuntime.Name != runtime.Name {
		return errors.New("the current backup file does not match the application")
	}

	if !isRollback {
		rollbackFile := path.Join(global.CONF.System.TmpDir, fmt.Sprintf("runtime/%s_%s.tar.gz", runtime.Name, time.Now().Format(constant.DateTimeSlimLayout)))
		if err := handleRuntimeBackup(runtime, path.Dir(rollbackFile), path.Base(rollbackFile), "", secret); err != nil {
			return fmt.Errorf("backup runtime %s for rollback before recover failed, err: %v", runtime.Name, err)
		}
		defer func() {
			if !isOk {
				global.LOG.Info("recover failed, start to rollback now")
				if err := handleRuntimeRecover(runtime, rollbackFile, true, secret); err != nil {
					global.LOG.Errorf("rollback runtime %s from %s failed, err: %v", runtime.Name, rollbackFile, err)
					return
				}
				global.LOG.Infof("rollback runtime %s from %s successful", runtime.Name, rollbackFile)
				_ = os.RemoveAll(rollbackFile)
			} else {
				_ = os.RemoveAll(rollbackFile)
			}
		}()
	}

	newEnvFile, err := coverEnvJsonToStr(runtime.Env)
	if err != nil {
		return err
	}
	runtimeDir := runtime.GetPath()
	backPath := fmt.Sprintf("%s_bak", runtimeDir)
	_ = fileOp.Rename(runtimeDir, backPath)
	_ = fileOp.CreateDir(runtimeDir, 0755)

	if err := handleUnTar(tmpPath+"/runtime.tar.gz", fmt.Sprintf("%s/%s", constant.RuntimeDir, runtime.Type), secret); err != nil {
		global.LOG.Errorf("handle recover from runtime.tar.gz failed, err: %v", err)
		_ = fileOp.DeleteDir(runtimeDir)
		_ = fileOp.Rename(backPath, runtimeDir)
		return err
	}
	_ = fileOp.DeleteDir(backPath)

	if len(newEnvFile) != 0 {
		envPath := fmt.Sprintf("%s/%s/%s/.env", constant.RuntimeDir, runtime.Type, runtime.Name)
		file, err := os.OpenFile(envPath, os.O_WRONLY|os.O_TRUNC, 0640)
		if err != nil {
			return err
		}
		defer file.Close()
		_, _ = file.WriteString(newEnvFile)
	}

	oldRuntime.ID = runtime.ID
	oldRuntime.Status = constant.RuntimeStarting
	if err := runtimeRepo.Save(&oldRuntime); err != nil {
		global.LOG.Errorf("save db app install failed, err: %v", err)
		return err
	}
	isOk = true
	return nil
}
