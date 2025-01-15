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
// @schemes http https

// @securityDefinitions.apikey CustomToken
// @description 自定义 Token 格式，格式：md5('1panel' + 1Panel-Token + 1Panel-Timestamp)。
// @description ```
// @description 示例请求头：
// @description curl -X GET "http://localhost:4004/api/v1/resource" \
// @description -H "1Panel-Token: <1panel_token>" \
// @description -H "1Panel-Timestamp: <current_unix_timestamp>"
// @description ```
// @description - `1Panel-Token` 为面板 API 接口密钥。
// @type apiKey
// @in Header
// @name 1Panel-Token
// @securityDefinitions.apikey Timestamp
// @type apiKey
// @in header
// @name 1Panel-Timestamp
// @description - `1Panel-Timestamp` 为当前时间的 Unix 时间戳（单位：秒）。

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
