package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(restoreCmd)
}

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "回滚1Panel",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("回滚成功")
		return nil
	},
}
