package main

import (
	"fmt"
	"os"

	_ "net/http/pprof"

	"github.com/1Panel-dev/1Panel/core/cmd/server/cmd"
	_ "github.com/1Panel-dev/1Panel/core/cmd/server/docs"
)

// @title 1Panel
// @version 1.0
// @description  Open Source Linux Panel
// @termsOfService http://swagger.io/terms/
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.apikey ApiKeyAuth
// @description Custom Token Format, Format: md5('1panel' + API-Key + UnixTimestamp).
// @description ```
// @description eg:
// @description curl -X GET "http://localhost:4004/api/v1/resource" \
// @description -H "1Panel-Token: <1panel_token>" \
// @description -H "1Panel-Timestamp: <current_unix_timestamp>"
// @description ```
// @description - `1Panel-Token` is the key for the panel API Key.
// @type apiKey
// @in Header
// @name 1Panel-Token
// @securityDefinitions.apikey Timestamp
// @type apiKey
// @in header
// @name 1Panel-Timestamp
// @description - `1Panel-Timestamp` is the Unix timestamp of the current time in seconds.

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
