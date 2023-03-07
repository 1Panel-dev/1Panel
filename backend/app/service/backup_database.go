package service

import (
	"compress/gzip"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/app/repo"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"github.com/pkg/errors"
)

func (u *BackupService) MysqlBackup(req dto.CommonBackup) error {
	localDir, err := loadLocalDir()
	if err != nil {
		return err
	}
	app, err := appInstallRepo.LoadBaseInfo("mysql", "")
	if err != nil {
		return err
	}

	timeNow := time.Now().Format("20060102150405")
	backupDir := fmt.Sprintf("%s/database/mysql/%s/%s", localDir, req.Name, req.DetailName)
	fileName := fmt.Sprintf("%s_%s.sql.gz", req.DetailName, timeNow)
	if err := handleMysqlBackup(app, backupDir, req.DetailName, fileName); err != nil {
		return err
	}
	record := &model.BackupRecord{
		Type:       "mysql",
		Name:       app.Name,
		DetailName: req.DetailName,
		Source:     "LOCAL",
		BackupType: "LOCAL",
		FileDir:    backupDir,
		FileName:   fileName,
	}
	if err := backupRepo.CreateRecord(record); err != nil {
		global.LOG.Errorf("save backup record failed, err: %v", err)
	}
	return nil
}

func (u *BackupService) MysqlRecover(req dto.CommonRecover) error {
	app, err := appInstallRepo.LoadBaseInfo("mysql", "")
	if err != nil {
		return err
	}
	fileOp := files.NewFileOp()
	if !fileOp.Stat(req.File) {
		return errors.New(fmt.Sprintf("%s file is not exist", req.File))
	}
	global.LOG.Infof("recover database %s-%s from backup file %s", req.Name, req.DetailName, req.File)
	if err := handleMysqlRecover(app, path.Dir(req.File), req.DetailName, path.Base(req.File), false); err != nil {
		return err
	}
	return nil
}

func (u *BackupService) MysqlRecoverByUpload(req dto.CommonRecover) error {
	app, err := appInstallRepo.LoadBaseInfo("mysql", "")
	if err != nil {
		return err
	}
	file := req.File
	fileName := path.Base(req.File)
	if strings.HasSuffix(fileName, ".tar.gz") {
		fileNameItem := time.Now().Format("20060102150405")
		dstDir := fmt.Sprintf("%s/%s", path.Dir(req.File), fileNameItem)
		if _, err := os.Stat(dstDir); err != nil && os.IsNotExist(err) {
			if err = os.MkdirAll(dstDir, os.ModePerm); err != nil {
				return fmt.Errorf("mkdir %s failed, err: %v", dstDir, err)
			}
		}
		if err := handleUnTar(req.File, dstDir); err != nil {
			_ = os.RemoveAll(dstDir)
			return err
		}
		global.LOG.Infof("decompress file %s successful, now start to check test.sql is exist", req.File)
		hasTestSql := false
		_ = filepath.Walk(dstDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() && info.Name() == "test.sql" {
				hasTestSql = true
				file = path
				fileName = "test.sql"
			}
			return nil
		})
		if !hasTestSql {
			_ = os.RemoveAll(dstDir)
			return fmt.Errorf("no such file named test.sql in %s, err: %v", fileName, err)
		}
		defer func() {
			_ = os.RemoveAll(dstDir)
		}()
	}

	if err := handleMysqlRecover(app, path.Dir(file), req.DetailName, fileName, false); err != nil {
		return err
	}
	global.LOG.Info("recover from uploads successful!")
	return nil
}

func handleMysqlBackup(app *repo.RootInfo, backupDir, dbName, fileName string) error {
	fileOp := files.NewFileOp()
	if !fileOp.Stat(backupDir) {
		if err := os.MkdirAll(backupDir, os.ModePerm); err != nil {
			return fmt.Errorf("mkdir %s failed, err: %v", backupDir, err)
		}
	}
	outfile, _ := os.OpenFile(backupDir+"/"+fileName, os.O_RDWR|os.O_CREATE, 0755)
	global.LOG.Infof("start to mysqldump | gzip > %s.gzip", backupDir+"/"+fileName)
	cmd := exec.Command("docker", "exec", app.ContainerName, "mysqldump", "-uroot", "-p"+app.Password, dbName)
	gzipCmd := exec.Command("gzip", "-cf")
	gzipCmd.Stdin, _ = cmd.StdoutPipe()
	gzipCmd.Stdout = outfile
	_ = gzipCmd.Start()
	_ = cmd.Run()
	_ = gzipCmd.Wait()

	return nil
}

func handleMysqlRecover(mysqlInfo *repo.RootInfo, recoverDir, dbName, fileName string, isRollback bool) error {
	isOk := false
	if !isRollback {
		rollbackFile := fmt.Sprintf("%s/original/database/%s_%s.sql.gz", global.CONF.System.BaseDir, mysqlInfo.Name, time.Now().Format("20060102150405"))
		if err := handleMysqlBackup(mysqlInfo, path.Dir(rollbackFile), dbName, path.Base(rollbackFile)); err != nil {
			return fmt.Errorf("backup mysql db %s for rollback before recover failed, err: %v", mysqlInfo.Name, err)
		}
		defer func() {
			if !isOk {
				if err := handleMysqlRecover(mysqlInfo, path.Dir(rollbackFile), dbName, path.Base(rollbackFile), true); err != nil {
					global.LOG.Errorf("rollback mysql db %s from %s failed, err: %v", dbName, rollbackFile, err)
					return
				}
				global.LOG.Infof("rollback mysql db %s from %s successful", dbName, rollbackFile)
				_ = os.RemoveAll(rollbackFile)
			} else {
				_ = os.RemoveAll(rollbackFile)
			}
		}()
	}
	file := recoverDir + "/" + fileName
	fi, _ := os.Open(file)
	defer fi.Close()
	cmd := exec.Command("docker", "exec", "-i", mysqlInfo.ContainerName, "mysql", "-uroot", "-p"+mysqlInfo.Password, dbName)
	if strings.HasSuffix(fileName, ".gz") {
		gzipFile, err := os.Open(file)
		if err != nil {
			return err
		}
		defer gzipFile.Close()
		gzipReader, err := gzip.NewReader(gzipFile)
		if err != nil {
			return err
		}
		defer gzipReader.Close()
		cmd.Stdin = gzipReader
	} else {
		cmd.Stdin = fi
	}
	stdout, err := cmd.CombinedOutput()
	stdStr := strings.ReplaceAll(string(stdout), "mysql: [Warning] Using a password on the command line interface can be insecure.\n", "")
	if err != nil || strings.HasPrefix(string(stdStr), "ERROR ") {
		return errors.New(stdStr)
	}
	isOk = true
	return nil
}
