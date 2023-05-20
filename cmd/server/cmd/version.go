package cmd

import (
	"fmt"

	"github.com/1Panel-dev/1Panel/backend/configs"
	"github.com/1Panel-dev/1Panel/cmd/server/conf"
	"gopkg.in/yaml.v3"

	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "获取系统版本信息",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := loadDBConn()
		if err != nil {
			return err
		}
		version := getSettingByKey(db, "SystemVersion")
		appStoreVersion := getSettingByKey(db, "AppStoreVersion")

		fmt.Printf("1panel version: %s\n", version)
		fmt.Printf("appstore version: %s\n", appStoreVersion)
		config := configs.ServerConfig{}
		if err := yaml.Unmarshal(conf.AppYaml, &config); err != nil {
			return fmt.Errorf("unmarshal conf.App.Yaml failed, errL %v", err)
		} else {
			fmt.Printf("mode: %s\n", config.System.Mode)
		}
		return nil
	},
}
