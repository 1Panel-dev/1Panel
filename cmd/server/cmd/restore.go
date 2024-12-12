package cmd

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/i18n"
	cmdUtils "github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/common"
	"github.com/pkg/errors"

	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(restoreCmd)
}

var restoreCmd = &cobra.Command{
	Use: "restore",
	RunE: func(cmd *cobra.Command, args []string) error {
		i18n.UseI18nForCmd(language)
		if !isRoot() {
			fmt.Println(i18n.GetMsgWithMapForCmd("SudoHelper", map[string]interface{}{"cmd": "sudo 1pctl restore"}))
			return nil
		}
		stdout, err := cmdUtils.Exec("grep '^BASE_DIR=' /usr/local/bin/1pctl | cut -d'=' -f2")
		if err != nil {
			return fmt.Errorf("handle load `BASE_DIR` failed, err: %v", err)
		}
		baseDir := strings.ReplaceAll(stdout, "\n", "")
		upgradeDir := path.Join(baseDir, "1panel", "tmp", "upgrade")

		tmpPath, err := loadRestorePath(upgradeDir)
		if err != nil {
			return err
		}
		if tmpPath == "no such file" {
			fmt.Println(i18n.GetMsgByKeyForCmd("RestoreNoSuchFile"))
			return nil
		}
		tmpPath = path.Join(upgradeDir, tmpPath, "original")
		fmt.Println(i18n.GetMsgWithMapForCmd("RestoreStep1", map[string]interface{}{"name": tmpPath}))

		if err := common.CopyFile(path.Join(tmpPath, "1panel"), "/usr/local/bin"); err != nil {
			return err
		}
		fmt.Println(i18n.GetMsgByKeyForCmd("RestoreStep2"))
		if err := common.CopyFile(path.Join(tmpPath, "1pctl"), "/usr/local/bin"); err != nil {
			return err
		}
		_, _ = cmdUtils.Execf("cp -r %s /usr/local/bin", path.Join(tmpPath, "lang"))
		geoPath := path.Join(global.CONF.System.BaseDir, "1panel/geo")
		_, _ = cmdUtils.Execf("mkdir %s && cp %s %s/", geoPath, path.Join(tmpPath, "GeoIP.mmdb"), geoPath)
		fmt.Println(i18n.GetMsgByKeyForCmd("RestoreStep3"))
		if err := common.CopyFile(path.Join(tmpPath, "1panel.service"), "/etc/systemd/system"); err != nil {
			return err
		}
		fmt.Println(i18n.GetMsgByKeyForCmd("RestoreStep4"))
		checkPointOfWal()
		if _, err := os.Stat(path.Join(tmpPath, "1Panel.db")); err == nil {
			if err := common.CopyFile(path.Join(tmpPath, "1Panel.db"), path.Join(baseDir, "1panel/db")); err != nil {
				return err
			}
		}
		if _, err := os.Stat(path.Join(tmpPath, "db.tar.gz")); err == nil {
			if err := handleUnTar(path.Join(tmpPath, "db.tar.gz"), path.Join(baseDir, "1panel")); err != nil {
				return err
			}
		}
		fmt.Println(i18n.GetMsgByKeyForCmd("RestoreStep5"))
		fmt.Println(i18n.GetMsgByKeyForCmd("RestoreSuccessful"))
		return nil
	},
}

func checkPointOfWal() {
	db, err := loadDBConn()
	if err != nil {
		return
	}
	_ = db.Exec("PRAGMA wal_checkpoint(TRUNCATE);").Error
}

func loadRestorePath(upgradeDir string) (string, error) {
	if _, err := os.Stat(upgradeDir); err != nil && os.IsNotExist(err) {
		return "no such file", nil
	}
	files, err := os.ReadDir(upgradeDir)
	if err != nil {
		return "", err
	}
	var folders []string
	for _, file := range files {
		if file.IsDir() {
			folders = append(folders, file.Name())
		}
	}
	if len(folders) == 0 {
		return "no such file", nil
	}
	sort.Slice(folders, func(i, j int) bool {
		return folders[i] > folders[j]
	})
	return folders[0], nil
}

func handleUnTar(sourceFile, targetDir string) error {
	if _, err := os.Stat(targetDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(targetDir, os.ModePerm); err != nil {
			return err
		}
	}

	commands := fmt.Sprintf("tar zxvfC %s %s", sourceFile, targetDir)
	stdout, err := cmdUtils.ExecWithTimeOut(commands, 20*time.Second)
	if err != nil {
		return errors.New(stdout)
	}
	return nil
}
