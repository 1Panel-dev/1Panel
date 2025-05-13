package cmd

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/i18n"
	cmdUtils "github.com/1Panel-dev/1Panel/core/utils/cmd"
	"github.com/1Panel-dev/1Panel/core/utils/common"
	"github.com/1Panel-dev/1Panel/core/utils/files"

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
		stdout, err := cmdUtils.RunDefaultWithStdoutBashC("grep '^BASE_DIR=' /usr/local/bin/1pctl | cut -d'=' -f2")
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
		if err := files.CopyItem(false, true, path.Join(tmpPath, "1panel-agent"), "/usr/local/bin"); err != nil {
			return err
		}
		if err := files.CopyItem(false, true, path.Join(tmpPath, "1panel-core"), "/usr/local/bin"); err != nil {
			return err
		}
		if err := files.CopyItem(true, true, path.Join(tmpPath, "lang"), "/usr/local/bin"); err != nil {
			return err
		}
		if err := files.CopyItem(false, true, path.Join(tmpPath, "GeoIP.mmdb"), path.Join(baseDir, "1panel/geo")); err != nil {
			return err
		}
		sudo := cmdUtils.SudoHandleCmd()
		_, _ = cmdUtils.RunDefaultWithStdoutBashCf("%s chmod 755 /usr/local/bin/1panel-agent /usr/local/bin/1panel-core", sudo)

		fmt.Println(i18n.GetMsgByKeyForCmd("RestoreStep2"))
		if err := files.CopyItem(false, true, path.Join(tmpPath, "1pctl"), "/usr/local/bin"); err != nil {
			return err
		}
		_, _ = cmdUtils.RunDefaultWithStdoutBashCf("%s chmod 755 /usr/local/bin/1pctl", sudo)
		_, _ = cmdUtils.RunDefaultWithStdoutBashCf("cp -r %s /usr/local/bin", path.Join(tmpPath, "lang"))
		geoPath := path.Join(global.CONF.Base.InstallDir, "1panel/geo")
		_, _ = cmdUtils.RunDefaultWithStdoutBashCf("mkdir %s && cp %s %s/", geoPath, path.Join(tmpPath, "GeoIP.mmdb"), geoPath)

		fmt.Println(i18n.GetMsgByKeyForCmd("RestoreStep3"))
		if err := files.CopyItem(false, true, path.Join(tmpPath, "1panel-core.service"), "/etc/systemd/system"); err != nil {
			return err
		}
		if err := files.CopyItem(false, true, path.Join(tmpPath, "1panel-agent.service"), "/etc/systemd/system"); err != nil {
			return err
		}
		fmt.Println(i18n.GetMsgByKeyForCmd("RestoreStep4"))
		if _, err := os.Stat(path.Join(tmpPath, "db")); err == nil {
			dbPath := path.Join(baseDir, "1panel")
			if err := files.CopyItem(true, true, path.Join(tmpPath, "db"), dbPath); err != nil {
				global.LOG.Errorf("rollback 1panel db failed, err: %v", err)
			}
		}

		fmt.Println(i18n.GetMsgByKeyForCmd("RestoreStep5"))
		fmt.Println(i18n.GetMsgByKeyForCmd("RestoreSuccessful"))
		return nil
	},
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
		return common.ComparePanelVersion(folders[i], folders[j])
	})
	return folders[0], nil
}
