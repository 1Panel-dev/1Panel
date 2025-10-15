package cmd

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/i18n"
	"github.com/1Panel-dev/1Panel/core/utils/cmd"
	"github.com/1Panel-dev/1Panel/core/utils/common"
	"github.com/1Panel-dev/1Panel/core/utils/encrypt"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func init() {
	updateCmd.SetHelpFunc(func(c *cobra.Command, s []string) {
		i18n.UseI18nForCmd(language)
		loadUpdateHelper()
	})

	RootCmd.AddCommand(updateCmd)
	updateCmd.AddCommand(updateUserName)
	updateCmd.AddCommand(updatePassword)
	updateCmd.AddCommand(updatePort)

	updateCmd.AddCommand(updateVersion)
}

var updateCmd = &cobra.Command{
	Use: "update",
	RunE: func(cmd *cobra.Command, args []string) error {
		i18n.UseI18nForCmd(language)
		loadUpdateHelper()
		return nil
	},
}

var updateUserName = &cobra.Command{
	Use:   "username",
	Short: i18n.GetMsgByKeyForCmd("UpdateUser"),
	RunE: func(cmd *cobra.Command, args []string) error {
		i18n.UseI18nForCmd(language)
		if !isRoot() {
			fmt.Println(i18n.GetMsgWithMapForCmd("SudoHelper", map[string]interface{}{"cmd": "sudo 1pctl update username"}))
			return nil
		}
		username()
		return nil
	},
}
var updatePassword = &cobra.Command{
	Use:   "password",
	Short: i18n.GetMsgByKeyForCmd("UpdatePassword"),
	RunE: func(cmd *cobra.Command, args []string) error {
		i18n.UseI18nForCmd(language)
		if !isRoot() {
			fmt.Println(i18n.GetMsgWithMapForCmd("SudoHelper", map[string]interface{}{"cmd": "sudo 1pctl update password"}))
			return nil
		}
		password()
		return nil
	},
}
var updatePort = &cobra.Command{
	Use:   "port",
	Short: i18n.GetMsgByKeyForCmd("UpdatePort"),
	RunE: func(cmd *cobra.Command, args []string) error {
		i18n.UseI18nForCmd(language)
		if !isRoot() {
			fmt.Println(i18n.GetMsgWithMapForCmd("SudoHelper", map[string]interface{}{"cmd": "sudo 1pctl update port"}))
			return nil
		}
		port()
		return nil
	},
}
var updateVersion = &cobra.Command{
	Use:  "version",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		i18n.UseI18nForCmd(language)
		if !isRoot() {
			fmt.Println(i18n.GetMsgWithMapForCmd("SudoHelper", map[string]interface{}{"cmd": "sudo 1pctl update version"}))
			return nil
		}
		version := args[0]
		if len(version) == 0 || !strings.HasPrefix(version, "v2.") {
			fmt.Println("err version in param input")
			return nil
		}
		db, err := loadDBConn("core.db")
		if err != nil {
			fmt.Println(i18n.GetMsgWithMapForCmd("DBConnErr", map[string]interface{}{"err": err.Error()}))
			return err
		}
		if err := setSettingByKey(db, "SystemVersion", version); err != nil {
			fmt.Println(i18n.GetMsgWithMapForCmd("UpdateUserErr", map[string]interface{}{"err": err.Error()}))
			return err
		}
		return nil
	},
}

func username() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(i18n.GetMsgByKeyForCmd("UpdateUser") + ": ")
	newUsername, _ := reader.ReadString('\n')
	newUsername = strings.Trim(newUsername, "\n")
	if len(newUsername) == 0 {
		fmt.Println(i18n.GetMsgByKeyForCmd("UpdateUserNull"))
		return
	}
	if strings.Contains(newUsername, " ") {
		fmt.Println(i18n.GetMsgByKeyForCmd("UpdateUserBlank"))
		return
	}
	result, err := regexp.MatchString("^[a-zA-Z0-9_\u4e00-\u9fa5]{3,30}$", newUsername)
	if !result || err != nil {
		fmt.Println(i18n.GetMsgByKeyForCmd("UpdateUserFormat"))
		return
	}

	db, err := loadDBConn("core.db")
	if err != nil {
		fmt.Println(i18n.GetMsgWithMapForCmd("DBConnErr", map[string]interface{}{"err": err.Error()}))
		return
	}
	if err := setSettingByKey(db, "UserName", newUsername); err != nil {
		fmt.Println(i18n.GetMsgWithMapForCmd("UpdateUserErr", map[string]interface{}{"err": err.Error()}))
		return
	}

	fmt.Println("\n" + i18n.GetMsgByKeyForCmd("UpdateSuccessful"))
	fmt.Println(i18n.GetMsgWithMapForCmd("UpdateUserResult", map[string]interface{}{"name": newUsername}))
}

func password() {
	fmt.Print(i18n.GetMsgByKeyForCmd("UpdatePassword") + ": ")
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println("\n" + i18n.GetMsgWithMapForCmd("UpdatePasswordRead", map[string]interface{}{"err": err.Error()}))
		return
	}
	newPassword := string(bytePassword)
	newPassword = strings.Trim(newPassword, "\n")

	if len(newPassword) == 0 {
		fmt.Println("\n", i18n.GetMsgByKeyForCmd("UpdatePasswordNull"))
		return
	}
	if strings.Contains(newPassword, " ") {
		fmt.Println("\n" + i18n.GetMsgByKeyForCmd("UpdateUPasswordBlank"))
		return
	}
	db, err := loadDBConn("core.db")
	if err != nil {
		fmt.Println("\n" + i18n.GetMsgWithMapForCmd("DBConnErr", map[string]interface{}{"err": err.Error()}))
		return
	}
	complexSetting := getSettingByKey(db, "ComplexityVerification")
	if complexSetting == constant.StatusEnable {
		if isValidPassword("newPassword") {
			fmt.Println("\n" + i18n.GetMsgByKeyForCmd("UpdatePasswordFormat"))
			return
		}
	}
	if len(newPassword) < 6 {
		fmt.Println(i18n.GetMsgByKeyForCmd("UpdatePasswordLen"))
		return
	}

	fmt.Print("\n" + i18n.GetMsgByKeyForCmd("UpdatePasswordRe"))
	byteConfirmPassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println("\n" + i18n.GetMsgWithMapForCmd("UpdatePasswordRead", map[string]interface{}{"err": err.Error()}))
		return
	}
	confirmPassword := string(byteConfirmPassword)
	confirmPassword = strings.Trim(confirmPassword, "\n")

	if newPassword != confirmPassword {
		fmt.Println("\n", i18n.GetMsgByKeyForCmd("UpdatePasswordSame"))
		return
	}

	p := ""
	encryptSetting := getSettingByKey(db, "EncryptKey")
	if len(encryptSetting) == 16 {
		global.CONF.Base.EncryptKey = encryptSetting
		p, _ = encrypt.StringEncrypt(newPassword)
	} else {
		p = newPassword
	}
	if err := setSettingByKey(db, "Password", p); err != nil {
		fmt.Println("\n", i18n.GetMsgWithMapForCmd("UpdatePortErr", map[string]interface{}{"err": err.Error()}))
		return
	}
	username := getSettingByKey(db, "UserName")

	fmt.Println("\n" + i18n.GetMsgByKeyForCmd("UpdateSuccessful"))
	fmt.Println(i18n.GetMsgWithMapForCmd("UpdateUserResult", map[string]interface{}{"name": username}))
	fmt.Println(i18n.GetMsgWithMapForCmd("UpdatePasswordResult", map[string]interface{}{"name": string(newPassword)}))
}

func port() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(i18n.GetMsgByKeyForCmd("UpdatePort") + ": ")

	newPortStr, _ := reader.ReadString('\n')
	newPortStr = strings.Trim(newPortStr, "\n")
	newPort, err := strconv.Atoi(strings.TrimSpace(newPortStr))
	if err != nil || newPort < 1 || newPort > 65535 {
		fmt.Println(i18n.GetMsgByKeyForCmd("UpdatePortFormat"))
		return
	}
	if common.ScanPort(newPort) {
		fmt.Println(i18n.GetMsgByKeyForCmd("UpdatePortUsed"))
		return
	}
	db, err := loadDBConn("core.db")
	if err != nil {
		fmt.Println(i18n.GetMsgWithMapForCmd("DBConnErr", map[string]interface{}{"err": err.Error()}))
		return
	}
	if err := setSettingByKey(db, "ServerPort", newPortStr); err != nil {
		fmt.Println(i18n.GetMsgWithMapForCmd("UpdatePortErr", map[string]interface{}{"err": err.Error()}))
		return
	}

	fmt.Println("\n" + i18n.GetMsgByKeyForCmd("UpdateSuccessful"))
	fmt.Println(i18n.GetMsgWithMapForCmd("UpdatePortResult", map[string]interface{}{"name": newPortStr}))

	std, err := cmd.RunDefaultWithStdoutBashC("1pctl restart core")
	if err != nil {
		fmt.Println(std)
	}
}
func isValidPassword(password string) bool {
	numCount := 0
	alphaCount := 0
	specialCount := 0

	for _, char := range password {
		switch {
		case unicode.IsDigit(char):
			numCount++
		case unicode.IsLetter(char):
			alphaCount++
		case isSpecialChar(char):
			specialCount++
		}
	}

	if len(password) < 8 && len(password) > 30 {
		return false
	}
	if (numCount == 0 && alphaCount == 0) || (alphaCount == 0 && specialCount == 0) || (numCount == 0 && specialCount == 0) {
		return false
	}
	return true
}

func isSpecialChar(char rune) bool {
	specialChars := "!@#$%*_,.?"
	return unicode.IsPunct(char) && contains(specialChars, char)
}

func contains(specialChars string, char rune) bool {
	for _, c := range specialChars {
		if c == char {
			return true
		}
	}
	return false
}

func loadUpdateHelper() {
	fmt.Println(i18n.GetMsgByKeyForCmd("UpdateCommands"))
	fmt.Println("\nUsage:\n  1panel update [command]\n\nAvailable Commands:")
	fmt.Println("\n  password    " + i18n.GetMsgByKeyForCmd("UpdatePassword"))
	fmt.Println("  port        " + i18n.GetMsgByKeyForCmd("UpdatePort"))
	fmt.Println("  username    " + i18n.GetMsgByKeyForCmd("UpdateUser"))
	fmt.Println("\nFlags:\n  -h, --help   help for update")
	fmt.Println("\nUse \"1panel update [command] --help\" for more information about a command.")
}
