package cmd

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	cmdUtils "github.com/1Panel-dev/1Panel/core/utils/cmd"
	"github.com/1Panel-dev/1Panel/core/utils/files"

	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(restoreCmd)
}

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "回滚 1Panel 服务及数据",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !isRoot() {
			fmt.Println("请使用 sudo 1pctl restore 或者切换到 root 用户")
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
		if tmpPath == "暂无可回滚文件" {
			fmt.Println("暂无可回滚文件")
			return nil
		}
		tmpPath = path.Join(upgradeDir, tmpPath, "original")
		fmt.Printf("(0/4) 开始从 %s 目录回滚 1Panel 服务及数据... \n", tmpPath)

		if err := files.CopyItem(false, true, path.Join(tmpPath, "1panel*"), "/usr/local/bin"); err != nil {
			return err
		}
		fmt.Println("(1/4) 1panel 二进制回滚成功")
		if err := files.CopyItem(false, true, path.Join(tmpPath, "1pctl"), "/usr/local/bin"); err != nil {
			return err
		}
		fmt.Println("(2/4) 1panel 脚本回滚成功")
		if err := files.CopyItem(false, true, path.Join(tmpPath, "1panel*.service"), "/etc/systemd/system"); err != nil {
			return err
		}
		fmt.Println("(3/4) 1panel 服务回滚成功")
		if _, err := os.Stat(path.Join(tmpPath, "core.db")); err == nil {
			if err := files.CopyItem(true, true, path.Join(tmpPath, "db"), path.Join(baseDir, "1panel")); err != nil {
				return err
			}
		}
		fmt.Printf("(4/4) 1panel 数据回滚成功 \n\n")

		fmt.Println("回滚成功！正在重启服务，请稍候...")
		return nil
	},
}

func loadRestorePath(upgradeDir string) (string, error) {
	if _, err := os.Stat(upgradeDir); err != nil && os.IsNotExist(err) {
		return "暂无可回滚文件", nil
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
		return "暂无可回滚文件", nil
	}
	sort.Slice(folders, func(i, j int) bool {
		return folders[i] > folders[j]
	})
	return folders[0], nil
}
