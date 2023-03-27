package client

import (
	"fmt"
	"strings"
	"testing"

	"github.com/1Panel-dev/1Panel/backend/utils/ssh"
)

func TestFire(t *testing.T) {
	ConnInfo := ssh.ConnInfo{
		Addr:     "172.16.10.234",
		User:     "ubuntu",
		AuthMode: "password",
		Port:     22,
	}
	output, err := ConnInfo.Run("sudo ufw status verbose")
	if err != nil {
		fmt.Println(err)
	}

	lines := strings.Split(string(output), "\n")
	var datas []FireInfo
	isStart := false
	for _, line := range lines {
		if strings.HasPrefix(line, "--") {
			isStart = true
			continue
		}
		if !isStart {
			continue
		}

	}
	fmt.Println(datas)
}
