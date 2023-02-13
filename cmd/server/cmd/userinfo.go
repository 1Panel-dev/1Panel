package cmd

import (
	"fmt"

	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/encrypt"
	"github.com/spf13/cobra"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init() {
	RootCmd.AddCommand(userinfoCmd)
}

var userinfoCmd = &cobra.Command{
	Use:   "userinfo",
	Short: "获取用户信息",
	RunE: func(cmd *cobra.Command, args []string) error {
		fullPath := "/opt/1panel/db/1Panel.db"
		db, err := gorm.Open(sqlite.Open(fullPath), &gorm.Config{})
		if err != nil {
			fmt.Printf("init my db conn failed, err: %v \n", err)
		}
		user := getSettingByKey(db, "UserName")
		password := getSettingByKey(db, "Password")
		port := getSettingByKey(db, "ServerPort")
		enptrySetting := getSettingByKey(db, "ServerPort")

		p := ""
		if len(enptrySetting) == 16 {
			global.CONF.System.EncryptKey = enptrySetting
			p, _ = encrypt.StringDecrypt(password)
		} else {
			p = password
		}

		fmt.Printf("username: %s\n", user)
		fmt.Printf("password: %s\n", p)
		fmt.Printf("port: %s\n", port)
		return nil
	},
}
