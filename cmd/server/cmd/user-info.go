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
	Short: "获取面板信息",
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
		pass := "********"
		if isDefault(db) {
			encryptSetting := getSettingByKey(db, "EncryptKey")
			pass = getSettingByKey(db, "Password")
			if len(encryptSetting) == 16 {
				global.CONF.System.EncryptKey = encryptSetting
				pass, _ = encrypt.StringDecrypt(pass)
			}
		}
		port := getSettingByKey(db, "ServerPort")
		ssl := getSettingByKey(db, "SSL")
		entrance := getSettingByKey(db, "SecurityEntrance")
		address := getSettingByKey(db, "SystemIP")

		protocol := "http"
		if ssl == "enable" {
			protocol = "https"
		}
		if address == "" {
			address = "$LOCAL_IP"
		}

		fmt.Printf("面板地址: %s://%s:%s/%s \n", protocol, address, port, entrance)
		fmt.Println("面板用户: ", user)
		fmt.Println("面板密码: ", pass)
		fmt.Println("提示：修改密码可执行命令：1pctl update password")
		return nil
	},
}
