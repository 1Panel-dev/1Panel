package cmd

import (
	"fmt"

	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/i18n"
	"github.com/spf13/cobra"
)

func init() {
	listenCmd.SetHelpFunc(func(c *cobra.Command, s []string) {
		i18n.UseI18nForCmd(language)
		loadListenIPHelper()
	})

	RootCmd.AddCommand(listenCmd)
	listenCmd.AddCommand(listenIpv4Cmd)
	listenCmd.AddCommand(listenIpv6Cmd)
}

var listenCmd = &cobra.Command{
	Use: "listen-ip",
	RunE: func(cmd *cobra.Command, args []string) error {
		i18n.UseI18nForCmd(language)
		loadListenIPHelper()
		return nil
	},
}

var listenIpv4Cmd = &cobra.Command{
	Use: "ipv4",
	RunE: func(cmd *cobra.Command, args []string) error {
		i18n.UseI18nForCmd(language)
		return updateBindInfo("ipv4")
	},
}
var listenIpv6Cmd = &cobra.Command{
	Use: "ipv6",
	RunE: func(cmd *cobra.Command, args []string) error {
		i18n.UseI18nForCmd(language)
		return updateBindInfo("ipv6")
	},
}

func updateBindInfo(protocol string) error {
	if !isRoot() {
		fmt.Println(i18n.GetMsgWithMapForCmd("SudoHelper", map[string]interface{}{"cmd": "sudo 1pctl listen-ip ipv6"}))
		return nil
	}
	db, err := loadDBConn("core.db")
	if err != nil {
		return err
	}
	ipv6 := constant.StatusDisable
	tcp := "tcp4"
	address := "0.0.0.0"
	if protocol == "ipv6" {
		ipv6 = constant.StatusEnable
		tcp = "tcp6"
		address = "::"
	}
	if err := setSettingByKey(db, "Ipv6", ipv6); err != nil {
		return err
	}
	if err := setSettingByKey(db, "BindAddress", address); err != nil {
		return err
	}
	fmt.Println(i18n.GetMsgWithMapForCmd("ListenChangeSuccessful", map[string]interface{}{"value": fmt.Sprintf(" %s [%s]", tcp, address)}))
	return nil
}

func loadListenIPHelper() {
	fmt.Println(i18n.GetMsgByKeyForCmd("UpdateCommands"))
	fmt.Println("\nUsage:\n  1panel listen-ip [command]\n\nAvailable Commands:")
	fmt.Println("\n  ipv4        " + i18n.GetMsgByKeyForCmd("ListenIPv4"))
	fmt.Println("  ipv6        " + i18n.GetMsgByKeyForCmd("ListenIPv6"))
	fmt.Println("\nFlags:\n  -h, --help   help for listen-ip")
	fmt.Println("\nUse \"1panel listen-ip [command] --help\" for more information about a command.")
}
