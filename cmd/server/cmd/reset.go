package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(resetCmd)
	resetCmd.AddCommand(resetMFACmd)
	resetCmd.AddCommand(resetSSLCmd)
	resetCmd.AddCommand(resetEntranceCmd)
	resetCmd.AddCommand(resetBindIpsCmd)
	resetCmd.AddCommand(resetDomainCmd)
}

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "重置系统信息",
}

var resetMFACmd = &cobra.Command{
	Use:   "mfa",
	Short: "取消 1Panel 两步验证",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := loadDBConn()
		if err != nil {
			return err
		}

		return setSettingByKey(db, "MFAStatus", "disable")
	},
}
var resetSSLCmd = &cobra.Command{
	Use:   "https",
	Short: "取消 1Panel https 方式登录",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := loadDBConn()
		if err != nil {
			return err
		}

		return setSettingByKey(db, "SSL", "disable")
	},
}
var resetEntranceCmd = &cobra.Command{
	Use:   "entrance",
	Short: "取消 1Panel 安全入口",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := loadDBConn()
		if err != nil {
			return err
		}

		return setSettingByKey(db, "SecurityEntrance", "")
	},
}
var resetBindIpsCmd = &cobra.Command{
	Use:   "ips",
	Short: "取消 1Panel 授权 IP 限制",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := loadDBConn()
		if err != nil {
			return err
		}

		return setSettingByKey(db, "AllowIPs", "")
	},
}
var resetDomainCmd = &cobra.Command{
	Use:   "domain",
	Short: "取消 1Panel 访问域名绑定",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := loadDBConn()
		if err != nil {
			return err
		}

		return setSettingByKey(db, "BindDomain", "")
	},
}
