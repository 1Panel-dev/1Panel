package service

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
)

func TestMysql(t *testing.T) {
	cmd := exec.Command("docker", "exec", "1Panel-redis-7.0.5-zgVH-K859", "redis-cli", "config", "get", "save")
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(stdout))
	}

	rows := strings.Split(string(stdout), "\r\n")
	rowMap := make(map[string]string)
	for _, v := range rows {
		itemRow := strings.Split(v, "\n")
		if len(itemRow) == 3 {
			rowMap[itemRow[0]] = itemRow[1]
		}
	}
	var info dto.RedisStatus
	arr, err := json.Marshal(rowMap)
	if err != nil {
		fmt.Println(err)
	}
	_ = json.Unmarshal(arr, &info)
	fmt.Println(info)
}
