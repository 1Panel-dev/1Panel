package main

import (
	"fmt"
	"os"

	"github.com/1Panel-dev/1Panel/cmd/server/cmd"
	_ "github.com/1Panel-dev/1Panel/cmd/server/docs"
	_ "net/http/pprof"
)

// @title 1Panel
// @version 1.0
// @description Top-Rated Web-based Linux Server Management Tool
// @termsOfService http://swagger.io/terms/
// @license.name GPL-3.0
// @license.url https://www.gnu.org/licenses/gpl-3.0.html
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.apikey CustomToken
// @description Custom Token Format, Format: md5('1panel' + 1Panel-Token + 1Panel-Timestamp).
// @description ```
// @description eg:
// @description curl -X GET "http://localhost:4004/api/v1/resource" \
// @description -H "1Panel-Token: <1panel_token>" \
// @description -H "1Panel-Timestamp: <current_unix_timestamp>"
// @description ```
// @description - `1Panel-Token` is the key for the panel API interface.
// @type apiKey
// @in Header
// @name 1Panel-Token
// @securityDefinitions.apikey Timestamp
// @type apiKey
// @in header
// @name 1Panel-Timestamp
// @description - `1Panel-Timestamp` is the Unix timestamp of the current time in seconds.

//go:generate swag init -o ./docs -g main.go -d ../../backend -g ../cmd/server/main.go
func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
