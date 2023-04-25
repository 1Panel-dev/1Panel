package cmd

import (
	"fmt"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/global"
	cmdUtils "github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/encrypt"
	"github.com/spf13/cobra"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init() {
	RootCmd.AddCommand(userinfoCmd)
}

var userinfoCmd = &cobra.Command{
	Use:   "user-info",
	Short: "获取用户信息",
	RunE: func(cmd *cobra.Command, args []string) error {
		stdout, err := cmdUtils.Exec("grep '^BASE_DIR=' /usr/bin/1pctl | cut -d'=' -f2")
		if err != nil {
			panic(err)
		}
		baseDir := strings.ReplaceAll(stdout, "\n", "")
		if len(baseDir) == 0 {
			fmt.Printf("error `BASE_DIR` find in /usr/bin/1pctl \n")
		}
		if strings.HasSuffix(baseDir, "/") {
			baseDir = baseDir[:strings.LastIndex(baseDir, "/")]
		}

		db, err := gorm.Open(sqlite.Open(baseDir+"/1panel/db/1Panel.db"), &gorm.Config{})
		if err != nil {
			fmt.Printf("init my db conn failed, err: %v \n", err)
		}
		user := getSettingByKey(db, "UserName")
		password := getSettingByKey(db, "Password")
		port := getSettingByKey(db, "ServerPort")
		ssl := getSettingByKey(db, "SSL")
		entrance := getSettingByKey(db, "SecurityEntrance")
		enptrySetting := getSettingByKey(db, "EncryptKey")

		p := ""
		if len(enptrySetting) == 16 {
			global.CONF.System.EncryptKey = enptrySetting
			p, _ = encrypt.StringDecrypt(password)
		} else {
			p = password
		}

		protocol := "http"
		if ssl == "enable" {
			protocol = "https"
		}

		fmt.Printf("username: %s\n", user)
		fmt.Printf("password: %s\n", p)
		fmt.Printf("port: %s\n", port)
		fmt.Printf("protocol: %s\n", protocol)
		fmt.Printf("entrance: %s\n", entrance)
		return nil
	},
}
