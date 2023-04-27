package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(resetMFACmd)
	RootCmd.AddCommand(resetSSLCmd)
	RootCmd.AddCommand(resetEntranceCmd)
}

var resetMFACmd = &cobra.Command{
	Use:   "reset-mfa",
	Short: "关闭 1Panel 两步验证",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := loadDBConn()
		if err != nil {
			return err
		}

		return setSettingByKey(db, "MFAStatus", "disable")
	},
}

var resetSSLCmd = &cobra.Command{
	Use:   "reset-https",
	Short: "取消 1Panel  https 方式登录",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := loadDBConn()
		if err != nil {
			return err
		}

		return setSettingByKey(db, "SSL", "disable")
	},
}
var resetEntranceCmd = &cobra.Command{
	Use:   "reset-entrance",
	Short: "取消 1Panel 安全入口",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := loadDBConn()
		if err != nil {
			return err
		}

		return setSettingByKey(db, "SecurityEntrance", "")
	},
}
