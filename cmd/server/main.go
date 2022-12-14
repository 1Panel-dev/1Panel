package main

import (
	"fmt"
	"github.com/1Panel-dev/1Panel/cmd/server/cmd"
	_ "github.com/1Panel-dev/1Panel/cmd/server/docs"
	"os"
)

// @title 1Panel
// @version 1.0
// @description  开源Linux面板
// @termsOfService http://swagger.io/terms/

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost
// @BasePath /api/v1

//go:generate swag init -o ./docs -d ../../backend/app/api/v1 -g ../../../../cmd/server/main.go
func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
