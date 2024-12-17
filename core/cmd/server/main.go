package main

import (
	"fmt"
	"os"

	_ "net/http/pprof"

	"github.com/1Panel-dev/1Panel/core/cmd/server/cmd"
	_ "github.com/1Panel-dev/1Panel/core/cmd/server/docs"
)

// @title 1Panel
// @version 2.0
// @description  开源Linux面板
// @termsOfService http://swagger.io/terms/
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost
// @BasePath /api/v2

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
