package cmd

import (
	"fmt"

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
		fullPath := "/opt/1panel/db/1Panel.db"
		db, err := gorm.Open(sqlite.Open(fullPath), &gorm.Config{})
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
