package cmd

import (
	"fmt"

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
		if !isRoot() {
			fmt.Println("请使用 sudo 1pctl reset mfa 或者切换到 root 用户")
			return nil
		}
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
		if !isRoot() {
			fmt.Println("请使用 sudo 1pctl reset https 或者切换到 root 用户")
			return nil
		}
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
		if !isRoot() {
			fmt.Println("请使用 sudo 1pctl reset entrance 或者切换到 root 用户")
			return nil
		}
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
		if !isRoot() {
			fmt.Println("请使用 sudo 1pctl reset ips 或者切换到 root 用户")
			return nil
		}
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
		if !isRoot() {
			fmt.Println("请使用 sudo 1pctl reset domain 或者切换到 root 用户")
			return nil
		}
		db, err := loadDBConn()
		if err != nil {
			return err
		}

		return setSettingByKey(db, "BindDomain", "")
	},
}
