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

		p := ""
		if len(encryptSetting) == 16 {
			global.CONF.System.EncryptKey = encryptSetting
			p, _ = encrypt.StringDecrypt(password)
		} else {
			p = password
		}

		fmt.Printf("username: %s\n", user)
		fmt.Printf("password: %s\n", p)
		fmt.Printf("port: %s\n", port)
		fmt.Printf("ssl: %s\n", ssl)
		fmt.Printf("entrance: %s\n", entrance)
		return nil
	},
}
