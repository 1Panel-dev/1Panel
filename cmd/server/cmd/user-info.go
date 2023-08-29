package cmd

import (
	"fmt"

	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/encrypt"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(userinfoCmd)
}

var userinfoCmd = &cobra.Command{
	Use:   "user-info",
	Short: "获取用户信息",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !isRoot() {
			fmt.Println("请使用 sudo 1pctl user-info 或者切换到 root 用户")
			return nil
		}
		db, err := loadDBConn()
		if err != nil {
			return fmt.Errorf("init my db conn failed, err: %v \n", err)
		}
		user := getSettingByKey(db, "UserName")
		password := getSettingByKey(db, "Password")
		port := getSettingByKey(db, "ServerPort")
		ssl := getSettingByKey(db, "SSL")
		entrance := getSettingByKey(db, "SecurityEntrance")
		encryptSetting := getSettingByKey(db, "EncryptKey")
		address := getSettingByKey(db, "SystemIP")

		p := ""
		if len(encryptSetting) == 16 {
			global.CONF.System.EncryptKey = encryptSetting
			p, _ = encrypt.StringDecrypt(password)
		} else {
			p = password
		}

		protocol := "http"
		if ssl == "enable" {
			protocol = "https"
		}
		if address == "" {
			address = "$LOCAL_IP"
		}

		fmt.Printf("面板地址: %s://%s:%s/%s \n", protocol, address, port, entrance)
		fmt.Printf("用户名称: %s\n", user)
		fmt.Printf("用户密码: %s\n", p)
		return nil
	},
}
