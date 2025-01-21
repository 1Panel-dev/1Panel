package cmd

import (
	"fmt"

	"github.com/1Panel-dev/1Panel/core/i18n"
	"github.com/spf13/cobra"
)

func init() {
	resetCmd.SetHelpFunc(func(c *cobra.Command, s []string) {
		i18n.UseI18nForCmd(language)
		loadResetHelper()
	})

	RootCmd.AddCommand(resetCmd)
	resetCmd.AddCommand(resetMFACmd)
	resetCmd.AddCommand(resetSSLCmd)
	resetCmd.AddCommand(resetEntranceCmd)
	resetCmd.AddCommand(resetBindIpsCmd)
	resetCmd.AddCommand(resetDomainCmd)
}

var resetCmd = &cobra.Command{
	Use: "reset",
	RunE: func(cmd *cobra.Command, args []string) error {
		i18n.UseI18nForCmd(language)
		loadResetHelper()
		return nil
	},
}

var resetMFACmd = &cobra.Command{
	Use: "mfa",
	RunE: func(cmd *cobra.Command, args []string) error {
		i18n.UseI18nForCmd(language)
		if !isRoot() {
			fmt.Println(i18n.GetMsgWithMapForCmd("SudoHelper", map[string]interface{}{"cmd": "sudo 1pctl reset mfa"}))
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
	Use: "https",
	RunE: func(cmd *cobra.Command, args []string) error {
		i18n.UseI18nForCmd(language)
		if !isRoot() {
			fmt.Println(i18n.GetMsgWithMapForCmd("SudoHelper", map[string]interface{}{"cmd": "sudo 1pctl reset https"}))
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
	Use: "entrance",
	RunE: func(cmd *cobra.Command, args []string) error {
		i18n.UseI18nForCmd(language)
		if !isRoot() {
			fmt.Println(i18n.GetMsgWithMapForCmd("SudoHelper", map[string]interface{}{"cmd": "sudo 1pctl reset entrance"}))
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
	Use: "ips",
	RunE: func(cmd *cobra.Command, args []string) error {
		i18n.UseI18nForCmd(language)
		if !isRoot() {
			fmt.Println(i18n.GetMsgWithMapForCmd("SudoHelper", map[string]interface{}{"cmd": "sudo 1pctl reset ips"}))
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
	Use: "domain",
	RunE: func(cmd *cobra.Command, args []string) error {
		i18n.UseI18nForCmd(language)
		if !isRoot() {
			fmt.Println(i18n.GetMsgWithMapForCmd("SudoHelper", map[string]interface{}{"cmd": "sudo 1pctl reset domain"}))
			return nil
		}
		db, err := loadDBConn()
		if err != nil {
			return err
		}

		return setSettingByKey(db, "BindDomain", "")
	},
}

func loadResetHelper() {
	fmt.Println(i18n.GetMsgByKeyForCmd("ResetCommands"))
	fmt.Println("\nUsage:\n  1panel reset [command]\n\nAvailable Commands:")
	fmt.Println("\n  domain      " + i18n.GetMsgByKeyForCmd("ResetDomain"))
	fmt.Println("  entrance    " + i18n.GetMsgByKeyForCmd("ResetEntrance"))
	fmt.Println("  https       " + i18n.GetMsgByKeyForCmd("ResetHttps"))
	fmt.Println("  ips         " + i18n.GetMsgByKeyForCmd("ResetIPs"))
	fmt.Println("  mfa         " + i18n.GetMsgByKeyForCmd("ResetMFA"))
	fmt.Println("\nFlags:\n  -h, --help   help for reset")
	fmt.Println("\nUse \"1panel reset [command] --help\" for more information about a command.")
}
