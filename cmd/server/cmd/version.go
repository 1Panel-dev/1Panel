package cmd

import (
	"fmt"
	"strings"

	cmdUtils "github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/spf13/cobra"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init() {
	RootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "获取系统版本信息",
	RunE: func(cmd *cobra.Command, args []string) error {
		stdout, err := cmdUtils.Exec("grep '^BASE_DIR=' /usr/bin/1pctl | cut -d'=' -f2")
		if err != nil {
			panic(err)
		}
		baseDir := strings.ReplaceAll(stdout, "\n", "")
		if len(baseDir) == 0 {
			fmt.Printf("error `BASE_DIR` find in /usr/bin/1pctl \n")
		}
		if strings.HasSuffix(baseDir, "/") {
			baseDir = baseDir[:strings.LastIndex(baseDir, "/")]
		}

		db, err := gorm.Open(sqlite.Open(baseDir+"/1panel/db/1Panel.db"), &gorm.Config{})
		if err != nil {
			fmt.Printf("init my db conn failed, err: %v \n", err)
		}
		version := getSettingByKey(db, "SystemVersion")
		appStoreVersion := getSettingByKey(db, "AppStoreVersion")

		fmt.Printf("1panel version: %s\n", version)
		fmt.Printf("appstore version: %s\n", appStoreVersion)
		return nil
	},
}
