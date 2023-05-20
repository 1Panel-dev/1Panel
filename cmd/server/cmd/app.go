package cmd

import (
	"bytes"
	"fmt"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"github.com/1Panel-dev/1Panel/cmd/server/app"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"io"
)

var (
	appKey     string
	appVersion string
)

func init() {
	initCmd.Flags().StringVarP(&appKey, "key", "k", "", "应用的key（仅支持英文）")
	initCmd.Flags().StringVarP(&appVersion, "version", "v", "", "应用版本")
	appCmd.AddCommand(initCmd)
	RootCmd.AddCommand(appCmd)
}

var appCmd = &cobra.Command{
	Use:   "app",
	Short: "应用相关命令",
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "初始化应用",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			appKey = args[0]
			if len(args) > 1 {
				appVersion = args[1]
			}
		}
		if appKey == "" {
			fmt.Println("应用的 key 缺失，使用 -k 指定")
			return nil
		}
		if appVersion == "" {
			fmt.Println("应用版本缺失，使用 -v 指定")
			return nil
		}
		fileOp := files.NewFileOp()
		appKeyPath := fmt.Sprintf("./%s", appKey)
		if err := createFolder(fileOp, appKeyPath); err != nil {
			return err
		}
		configYamlPath := fmt.Sprintf("%s/data.yml", appKeyPath)
		if err := createFile(fileOp, configYamlPath); err != nil {
			return err
		}
		if err := writeFile(fileOp, configYamlPath, bytes.NewReader(app.Config)); err != nil {
			return err
		}
		readMePath := fmt.Sprintf("%s/README.md", appKeyPath)
		if err := createFile(fileOp, readMePath); err != nil {
			return err
		}
		logoPath := fmt.Sprintf("%s/logo.png", appKeyPath)
		if err := createFile(fileOp, logoPath); err != nil {
			return err
		}
		if err := writeFile(fileOp, logoPath, bytes.NewReader(app.Logo)); err != nil {
			return err
		}
		versionPath := fmt.Sprintf("%s/%s", appKeyPath, appVersion)
		if fileOp.Stat(versionPath) {
			return errors.New("版本已存在！")
		}
		if err := createFolder(fileOp, versionPath); err != nil {
			return err
		}
		versionParamPath := fmt.Sprintf("%s/%s", versionPath, "data.yml")
		if err := createFile(fileOp, versionParamPath); err != nil {
			return err
		}
		if err := writeFile(fileOp, versionParamPath, bytes.NewReader(app.Param)); err != nil {
			return err
		}
		dockerComposeYamlPath := fmt.Sprintf("%s/%s", versionPath, "docker-compose.yml")
		if err := createFile(fileOp, dockerComposeYamlPath); err != nil {
			return err
		}
		fmt.Println("创建成功！")
		return nil
	},
}

func createFile(fileOp files.FileOp, filePath string) error {
	if fileOp.Stat(filePath) {
		return nil
	}
	if err := fileOp.CreateFile(filePath); err != nil {
		fmt.Printf("文件 %s 创建失败 %v", filePath, err)
		return err
	}
	return nil
}

func createFolder(fileOp files.FileOp, dirPath string) error {
	if fileOp.Stat(dirPath) {
		return nil
	}
	if err := fileOp.CreateDir(dirPath, 0755); err != nil {
		fmt.Printf("文件夹 %s 创建失败 %v", dirPath, err)
		return err
	}
	return nil
}

func writeFile(fileOp files.FileOp, filePath string, in io.Reader) error {
	if err := fileOp.WriteFile(filePath, in, 0755); err != nil {
		fmt.Printf("文件 %s 写入失败 %v", filePath, err)
		return err
	}
	return nil
}
