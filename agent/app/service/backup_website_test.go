package service

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/glebarez/sqlite"
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/language"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestWebsiteBackupWithLocalPHPRuntime(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "backup.db")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.Setting{}, &model.Runtime{}, &model.Website{}); err != nil {
		t.Fatal(err)
	}
	baseDir := t.TempDir()
	oldDB, oldDir, oldLog, oldI18n := global.DB, global.Dir, global.LOG, global.I18n
	global.DB = db
	global.Dir = global.SystemDir{DataDir: baseDir, RuntimeDir: filepath.Join(baseDir, "runtime")}
	global.LOG = logrus.New()
	global.LOG.SetOutput(io.Discard)
	global.I18n = goi18n.NewLocalizer(goi18n.NewBundle(language.English), "en")
	t.Cleanup(func() {
		global.DB, global.Dir, global.LOG, global.I18n = oldDB, oldDir, oldLog, oldI18n
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})

	// Creating any PHP runtime makes <runtime dir>/php, but a local runtime gets no directory of its own.
	if err := os.MkdirAll(filepath.Join(global.Dir.RuntimeDir, constant.RuntimePHP), 0755); err != nil {
		t.Fatal(err)
	}
	runtime := model.Runtime{Name: "php81l", Type: constant.RuntimePHP, Resource: constant.ResourceLocal, Version: "8.1", Status: constant.StatusNormal}
	if err := db.Create(&runtime).Error; err != nil {
		t.Fatal(err)
	}
	website := model.Website{PrimaryDomain: "local-php.test", Alias: "local-php.test", Type: constant.Runtime, RuntimeID: runtime.ID}
	if err := db.Create(&website).Error; err != nil {
		t.Fatal(err)
	}

	confPath := GetWebsiteConfigPath(website)
	indexPath := filepath.Join(GetSitePath(website, SiteIndexDir), "index.php")
	for file, content := range map[string]string{confPath: "server {}\n", indexPath: "<?php echo 'ok';\n"} {
		if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	backupDir := t.TempDir()
	if err := doWebsiteBackup(&website, &task.Task{Logger: global.LOG}, backupDir, "local-php.test.tar.gz", "", ""); err != nil {
		t.Fatalf("backup of a website on a local PHP runtime failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(backupDir, "local-php.test.tar.gz")); err != nil {
		t.Fatal(err)
	}
}
