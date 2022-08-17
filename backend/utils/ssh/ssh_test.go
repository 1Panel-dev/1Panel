package ssh

import (
	"fmt"
	"testing"
)

func TestSSH(t *testing.T) {
	ss := ConnInfo{
		Addr:     "172.16.10.111",
		Port:     22,
		User:     "root",
		AuthMode: "password",
		Password: "Calong@2015",
	}
	_, err := ss.NewClient()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(ss.Run("ip a"))
}
