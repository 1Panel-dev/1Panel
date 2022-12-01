package nginx

import (
	"fmt"
	"testing"
)

func TestNginx(t *testing.T) {
	//config, err := GetConfig("/opt/1Panel/data/apps/nginx/nginx-1/conf/conf.d/word-1.conf")
	config, err := GetConfig("/opt/1Panel/data/apps/nginx/nginx-new/conf/conf.d/1panel.cloud.conf")
	if err != nil {
		panic(err)
	}

	//server := config.FindServers()[0]
	//fmt.Println(server)
	//serverD := config.FindServers()[0]
	//serverD.AddListen("8989", false)

	fmt.Println(DumpConfig(config, IndentedStyle))
}
