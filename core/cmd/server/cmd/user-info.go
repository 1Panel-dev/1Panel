package cmd

import (
	"fmt"

	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/i18n"
	"github.com/1Panel-dev/1Panel/core/utils/encrypt"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(userinfoCmd)
}

var userinfoCmd = &cobra.Command{
	Use: "user-info",
	RunE: func(cmd *cobra.Command, args []string) error {
		i18n.UseI18nForCmd(language)
		if !isRoot() {
			fmt.Println(i18n.GetMsgWithMapForCmd("SudoHelper", map[string]interface{}{"cmd": "sudo 1pctl user-info"}))
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
				global.CONF.Base.EncryptKey = encryptSetting
				pass, _ = encrypt.StringDecrypt(pass)
			}
		}
		port := getSettingByKey(db, "ServerPort")
		ssl := getSettingByKey(db, "SSL")
		entrance := getSettingByKey(db, "SecurityEntrance")
		address := getSettingByKey(db, "SystemIP")

		protocol := "http"
		if ssl == constant.StatusEnable {
			protocol = "https"
		}
		if address == "" {
			address = "$LOCAL_IP"
		}

		fmt.Println(i18n.GetMsgByKeyForCmd("UserInfoAddr") + fmt.Sprintf("%s://%s:%s/%s ", protocol, address, port, entrance))
		fmt.Println(i18n.GetMsgWithMapForCmd("UpdateUserResult", map[string]interface{}{"name": user}))
		fmt.Println(i18n.GetMsgWithMapForCmd("UpdatePasswordResult", map[string]interface{}{"name": pass}))
		fmt.Println(i18n.GetMsgByKeyForCmd("UserInfoPassHelp") + "1pctl update password")
		return nil
	},
}
