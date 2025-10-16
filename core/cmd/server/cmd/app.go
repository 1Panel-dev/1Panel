package cmd

import (
	"bytes"
	"fmt"
	"github.com/1Panel-dev/1Panel/core/cmd/server/app"
	"github.com/1Panel-dev/1Panel/core/i18n"
	"io"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	appKey     string
	appVersion string
)

func init() {
	appCmd.SetHelpFunc(func(c *cobra.Command, s []string) {
		i18n.UseI18nForCmd(language)
		loadAppHelper()
	})
	initCmd.SetHelpFunc(func(c *cobra.Command, s []string) {
		i18n.UseI18nForCmd(language)
		loadAppInitHelper()
	})
	initCmd.Flags().StringVarP(&appKey, "key", "k", "", "")
	initCmd.Flags().StringVarP(&appVersion, "version", "v", "", "")
	appCmd.AddCommand(initCmd)
	RootCmd.AddCommand(appCmd)
}

var appCmd = &cobra.Command{
	Use: "app",
}

var initCmd = &cobra.Command{
	Use: "init",
	RunE: func(cmd *cobra.Command, args []string) error {
		i18n.UseI18nForCmd(language)
		if !isRoot() {
			fmt.Println(i18n.GetMsgWithMapForCmd("SudoHelper", map[string]interface{}{"cmd": "sudo 1pctl app init"}))
			return nil
		}
		if len(args) > 0 {
			appKey = args[0]
			if len(args) > 1 {
				appVersion = args[1]
			}
		}
		if appKey == "" {
			fmt.Println(i18n.GetMsgByKeyForCmd("AppMissKey"))
			return nil
		}
		if appVersion == "" {
			fmt.Println(i18n.GetMsgByKeyForCmd("AppMissVersion"))
			return nil
		}
		appKeyPath := fmt.Sprintf("./%s", appKey)
		if err := createFolder(appKeyPath); err != nil {
			return err
		}
		configYamlPath := fmt.Sprintf("%s/data.yml", appKeyPath)
		if err := createFile(configYamlPath); err != nil {
			return err
		}
		if err := writeFile(configYamlPath, bytes.NewReader(app.Config)); err != nil {
			return err
		}
		readMePath := fmt.Sprintf("%s/README.md", appKeyPath)
		if err := createFile(readMePath); err != nil {
			return err
		}
		logoPath := fmt.Sprintf("%s/logo.png", appKeyPath)
		if err := createFile(logoPath); err != nil {
			return err
		}
		if err := writeFile(logoPath, bytes.NewReader(app.Logo)); err != nil {
			return err
		}
		versionPath := fmt.Sprintf("%s/%s", appKeyPath, appVersion)
		if _, err := os.Stat(versionPath); err == nil {
			return errors.New(i18n.GetMsgByKeyForCmd("AppVersionExist"))
		}
		if err := createFolder(versionPath); err != nil {
			return err
		}
		versionParamPath := fmt.Sprintf("%s/%s", versionPath, "data.yml")
		if err := createFile(versionParamPath); err != nil {
			return err
		}
		if err := writeFile(versionParamPath, bytes.NewReader(app.Param)); err != nil {
			return err
		}
		dockerComposeYamlPath := fmt.Sprintf("%s/%s", versionPath, "docker-compose.yml")
		if err := createFile(dockerComposeYamlPath); err != nil {
			return err
		}
		fmt.Println(i18n.GetMsgByKeyForCmd("AppCreateSuccessful"))
		return nil
	},
}

func createFile(filePath string) error {
	if _, err := os.Stat(filePath); err == nil {
		return nil
	}
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println(i18n.GetMsgWithMapForCmd("AppCreateFileErr", map[string]interface{}{"name": filePath, "err": err.Error()}))
		return err
	}
	defer file.Close()
	return nil
}

func createFolder(dirPath string) error {
	if _, err := os.Stat(dirPath); err == nil {
		return nil
	}
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		fmt.Println(i18n.GetMsgWithMapForCmd("AppCreateDirErr", map[string]interface{}{"name": dirPath, "err": err.Error()}))
		return err
	}
	return nil
}

func writeFile(filePath string, in io.Reader) error {
	data, err := io.ReadAll(in)
	if err != nil {
		fmt.Println(i18n.GetMsgWithMapForCmd("AppWriteErr", map[string]interface{}{"name": filePath, "err": err.Error()}))
		return err
	}
	if err := os.WriteFile(filePath, data, 0755); err != nil {
		fmt.Println(i18n.GetMsgWithMapForCmd("AppWriteErr", map[string]interface{}{"name": filePath, "err": err.Error()}))
		return err
	}
	return nil
}

func loadAppHelper() {
	fmt.Println(i18n.GetMsgByKeyForCmd("AppCommands"))
	fmt.Println("\nUsage:\n  1panel app [command]\n\nAvailable Commands:")
	fmt.Println("\n  init        " + i18n.GetMsgByKeyForCmd("AppInit"))
	fmt.Println("\nFlags:\n  -h, --help             help for app")
	fmt.Println("  -k, --key string       " + i18n.GetMsgByKeyForCmd("AppKeyVal"))
	fmt.Println("  -v, --version string   " + i18n.GetMsgByKeyForCmd("AppVersion"))
	fmt.Println("\nUse \"1panel app [command] --help\" for more information about a command.")
}

func loadAppInitHelper() {
	fmt.Println(i18n.GetMsgByKeyForCmd("AppInit"))
	fmt.Println("\nUsage:\n  1panel app init [flags]")
	fmt.Println("\nFlags:\n  -h, --help             help for app")
	fmt.Println("  -k, --key string       " + i18n.GetMsgByKeyForCmd("AppKeyVal"))
	fmt.Println("  -v, --version string   " + i18n.GetMsgByKeyForCmd("AppVersion"))
}
