package service

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
)

func TestSfq(t *testing.T) {
	data := dto.SSHInfo{
		Port:                   "22",
		ListenAddress:          "0.0.0.0",
		PasswordAuthentication: "yes",
		PubkeyAuthentication:   "yes",
		PermitRootLogin:        "yes",
		UseDNS:                 "yes",
	}
	sshConf, err := os.ReadFile("/Downloads/sshd_config")
	if err != nil {
		fmt.Println(err)
	}
	lines := strings.Split(string(sshConf), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Port ") {
			data.Port = strings.ReplaceAll(line, "Port ", "")
		}
		if strings.HasPrefix(line, "ListenAddress ") {
			data.ListenAddress = strings.ReplaceAll(line, "ListenAddress ", "")
		}
		if strings.HasPrefix(line, "PasswordAuthentication ") {
			data.PasswordAuthentication = strings.ReplaceAll(line, "PasswordAuthentication ", "")
		}
		if strings.HasPrefix(line, "PubkeyAuthentication ") {
			data.PubkeyAuthentication = strings.ReplaceAll(line, "PubkeyAuthentication ", "")
		}
		if strings.HasPrefix(line, "PermitRootLogin ") {
			data.PermitRootLogin = strings.ReplaceAll(line, "PermitRootLogin ", "")
		}
		if strings.HasPrefix(line, "UseDNS ") {
			data.UseDNS = strings.ReplaceAll(line, "UseDNS ", "")
		}
	}
	fmt.Println(data)
}
