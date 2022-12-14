package cmd

import (
	"github.com/1Panel-dev/1Panel/backend/server"
	"github.com/spf13/cobra"
)

var (
	configPath string
)

func init() {
	RootCmd.Flags().BoolP("run", "r", false, "运行面板")
	RootCmd.Flags().StringVarP(&configPath, "config", "c", "/opt/1panel/conf/app.yml", "配置文件路径")
}

var RootCmd = &cobra.Command{
	Use:   "1panel",
	Short: "1Panel ，一款现代化的 Linux 面板",
	Long: `欢迎使用 1Panel 面板
github地址: https://github.com/1Panel-dev/1Panel
你可以使用如下命令操作1Panel
例如: 1panel -r 启动1Panel服务
1panel backup -d /some/path  备份1Panel
你也可以使用 1panel --help 查看帮助信息
或者使用 1panel xx --help 查看具体命令的帮助信息`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 || args[0] == "run" || args[0] == "r" {
			server.Start()
			return nil
		}
		return nil
	},
}
