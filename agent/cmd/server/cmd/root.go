package cmd

import (
	"github.com/1Panel-dev/1Panel/agent/server"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use: "1panel-agent",
	RunE: func(cmd *cobra.Command, args []string) error {
		server.Start()
		return nil
	},
}
