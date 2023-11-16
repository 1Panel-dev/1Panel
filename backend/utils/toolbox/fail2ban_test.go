package toolbox

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/1Panel-dev/1Panel/backend/utils/ssh"
)

func TestCds(t *testing.T) {
	kk := ssh.ConnInfo{
		Port:     22,
		AuthMode: "password",
		User:     "root",
	}
	sd, err := kk.Run("fail2ban-client get sshd ignoreip")
	if err != nil {
		fmt.Println(err)
	}
	sd = strings.ReplaceAll(sd, "|", "")
	sd = strings.ReplaceAll(sd, "`", "")
	sd = strings.ReplaceAll(sd, "\n", "")

	addrs := strings.Split(sd, "-")
	for _, addr := range addrs {
		if !strings.HasPrefix(addr, " ") {
			continue
		}
		fmt.Println(strings.TrimPrefix(addr, " "))
	}
}

func TestCdsxx(t *testing.T) {
	initFile := `#DEFAULT-START
[DEFAULT]
ignoreip = 127.0.0.1/8,172.16.10.114,172.16.10.116
bantime = 600
findtime = 300
maxretry = 5
banaction = firewallcmd-ipset
action = %(action_mwl)s
#DEFAULT-END

#sshd-START
[sshd]
enabled = true
filter = sshd
port = 22
maxretry = 5
findtime = 300
bantime = 86400
action = %(action_mwl)s
logpath = /var/log/secure
#sshd-END`

	if err := os.WriteFile("/Users/slooop/Downloads/tex.txt", []byte(initFile), 0640); err != nil {
		fmt.Println(err)
	}
}
