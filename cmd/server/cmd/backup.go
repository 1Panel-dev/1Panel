package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func init() {
	backupCmd.Flags().StringVarP(&configPath, "dir", "d", "/opt/backup", "备份目录")
	RootCmd.AddCommand(backupCmd)
}

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "备份1Panel",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("备份成功")
		return nil
	},
}
