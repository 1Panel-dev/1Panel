package cmd

import (
	"fmt"

	"github.com/1Panel-dev/1Panel/backend/configs"
	"github.com/1Panel-dev/1Panel/backend/i18n"
	"github.com/1Panel-dev/1Panel/cmd/server/conf"
	"gopkg.in/yaml.v3"

	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use: "version",
	RunE: func(cmd *cobra.Command, args []string) error {
		i18n.UseI18nForCmd(language)
		if !isRoot() {
			fmt.Println(i18n.GetMsgWithMapForCmd("SudoHelper", map[string]interface{}{"cmd": "sudo 1pctl version"}))
			return nil
		}
		db, err := loadDBConn()
		if err != nil {
			return err
		}
		version := getSettingByKey(db, "SystemVersion")

		fmt.Println(i18n.GetMsgByKeyForCmd("SystemVersion") + version)
		config := configs.ServerConfig{}
		if err := yaml.Unmarshal(conf.AppYaml, &config); err != nil {
			return fmt.Errorf("unmarshal conf.App.Yaml failed, err: %v", err)
		} else {
			fmt.Println(i18n.GetMsgByKeyForCmd("SystemMode") + config.System.Mode)
		}
		return nil
	},
}
