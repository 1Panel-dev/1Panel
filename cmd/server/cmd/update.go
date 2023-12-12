package cmd

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/common"
	"github.com/1Panel-dev/1Panel/backend/utils/encrypt"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func init() {
	RootCmd.AddCommand(updateCmd)
	updateCmd.AddCommand(updateUserName)
	updateCmd.AddCommand(updatePassword)
	updateCmd.AddCommand(updatePort)
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "修改面板信息",
}

var updateUserName = &cobra.Command{
	Use:   "username",
	Short: "修改面板用户",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !isRoot() {
			fmt.Println("请使用 sudo 1pctl update username 或者切换到 root 用户")
			return nil
		}
		username()
		return nil
	},
}
var updatePassword = &cobra.Command{
	Use:   "password",
	Short: "修改面板密码",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !isRoot() {
			fmt.Println("请使用 sudo 1pctl update password 或者切换到 root 用户")
			return nil
		}
		password()
		return nil
	},
}
var updatePort = &cobra.Command{
	Use:   "port",
	Short: "修改面板端口",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !isRoot() {
			fmt.Println("请使用 sudo 1pctl update port 或者切换到 root 用户")
			return nil
		}
		port()
		return nil
	},
}

func username() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("修改面板用户: ")
	newUsername, _ := reader.ReadString('\n')
	newUsername = strings.Trim(newUsername, "\n")
	if len(newUsername) == 0 {
		fmt.Println("错误：输入面板用户为空！")
		return
	}
	if strings.Contains(newUsername, " ") {
		fmt.Println("错误：输入面板用户中包含空格字符！")
		return
	}
	result, err := regexp.MatchString("^[a-zA-Z0-9_\u4e00-\u9fa5]{3,30}$", newUsername)
	if !result || err != nil {
		fmt.Println("错误：输入面板用户错误！仅支持英文、中文、数字和_,长度3-30")
		return
	}

	db, err := loadDBConn()
	if err != nil {
		fmt.Printf("错误：初始化数据库连接失败，%v\n", err)
		return
	}
	if err := setSettingByKey(db, "UserName", newUsername); err != nil {
		fmt.Printf("错误：面板用户修改失败，%v\n", err)
		return
	}

	fmt.Printf("修改成功！\n\n")
	fmt.Printf("面板用户：%s\n", newUsername)
}

func password() {
	fmt.Print("修改面板密码：")
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Printf("\n错误：面板密码信息读取错误，%v\n", err)
		return
	}
	newPassword := string(bytePassword)
	newPassword = strings.Trim(newPassword, "\n")

	if len(newPassword) == 0 {
		fmt.Println("\n错误：输入面板密码为空！")
		return
	}
	if strings.Contains(newPassword, " ") {
		fmt.Println("\n错误：输入面板密码中包含空格字符！")
		return
	}
	db, err := loadDBConn()
	if err != nil {
		fmt.Printf("\n错误：初始化数据库连接失败，%v\n", err)
		return
	}
	complexSetting := getSettingByKey(db, "ComplexityVerification")
	if complexSetting == "enable" {
		if isValidPassword("newPassword") {
			fmt.Println("\n错误：面板密码仅支持字母、数字、特殊字符（!@#$%*_,.?），长度 8-30 位！")
			return
		}
	}
	if len(newPassword) < 6 {
		fmt.Println("错误：请输入 6 位以上密码！")
		return
	}

	fmt.Print("\n确认密码：")
	byteConfirmPassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Printf("\n错误：面板密码信息读取错误，%v\n", err)
		return
	}
	confirmPassword := string(byteConfirmPassword)
	confirmPassword = strings.Trim(confirmPassword, "\n")

	if newPassword != confirmPassword {
		fmt.Printf("\n错误：两次密码不匹配，请检查后重试！，%v\n", err)
		return
	}

	p := ""
	encryptSetting := getSettingByKey(db, "EncryptKey")
	if len(encryptSetting) == 16 {
		global.CONF.System.EncryptKey = encryptSetting
		p, _ = encrypt.StringEncrypt(newPassword)
	} else {
		p = newPassword
	}
	if err := setSettingByKey(db, "Password", p); err != nil {
		fmt.Printf("\n错误：面板密码修改失败，%v\n", err)
		return
	}
	username := getSettingByKey(db, "UserName")

	fmt.Printf("\n修改成功！\n\n")
	fmt.Printf("面板用户：%s\n", username)
	fmt.Printf("面板密码：%s\n", string(newPassword))
}

func port() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("修改面板端口：")

	newPortStr, _ := reader.ReadString('\n')
	newPortStr = strings.Trim(newPortStr, "\n")
	newPort, err := strconv.Atoi(strings.TrimSpace(newPortStr))
	if err != nil || newPort < 1 || newPort > 65535 {
		fmt.Println("错误：输入的端口号必须在 1 到 65535 之间！")
		return
	}
	if common.ScanPort(newPort) {
		fmt.Println("错误：该端口号正被占用，请检查后重试！")
		return
	}
	db, err := loadDBConn()
	if err != nil {
		fmt.Printf("错误：初始化数据库连接失败，%v\n", err)
		return
	}
	if err := setSettingByKey(db, "ServerPort", newPortStr); err != nil {
		fmt.Printf("错误：面板端口修改失败，%v\n", err)
		return
	}

	fmt.Printf("修改成功！\n\n")
	fmt.Printf("面板端口：%s\n", newPortStr)

	std, err := cmd.Exec("1pctl restart")
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
