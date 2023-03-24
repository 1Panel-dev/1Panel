package client

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/1Panel-dev/1Panel/backend/utils/ssh"
)

func TestFire(t *testing.T) {
	ConnInfo := ssh.ConnInfo{
		User:     "ubuntu",
		AuthMode: "password",
		Port:     22,
	}
	output, err := ConnInfo.Run("sudo ufw status numbered")
	if err != nil {
		fmt.Println(err)
	}

	lines := strings.Split(string(output), "\n")
	var rules []UfwRule
	for _, line := range lines {
		if line == "" || !strings.HasPrefix(line, "[") {
			continue
		}
		fields := strings.Fields(line)
		rule := UfwRule{Status: fields[0], From: fields[1], To: fields[2], Proto: fields[3], Comment: strings.Join(fields[4:], " ")}
		rules = append(rules, rule)
	}
	ufwStatus := UfwStatus{Rules: rules}
	ufwStatusJSON, err := json.MarshalIndent(ufwStatus, "", " ")
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Println(string(ufwStatusJSON))

}

type UfwRule struct {
	Status  string `json:"status"`
	From    string `json:"from"`
	To      string `json:"to"`
	Proto   string `json:"proto"`
	Comment string `json:"comment"`
}
type UfwStatus struct {
	Rules []UfwRule `json:"rules"`
}
