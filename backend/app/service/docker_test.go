package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
)

func TestDocker(t *testing.T) {
	file, err := ioutil.ReadFile("/opt/1Panel/docker/daemon.json")
	if err != nil {
		fmt.Println(err)
	}
	var conf daemonJsonItem
	deamonMap := make(map[string]interface{})
	if err := json.Unmarshal(file, &deamonMap); err != nil {
		fmt.Println(err)
	}
	arr, err := json.Marshal(deamonMap)
	if err != nil {
		fmt.Println(err)
	}
	_ = json.Unmarshal(arr, &conf)

	for _, opt := range conf.ExecOpts {
		if strings.HasPrefix(opt, "native.cgroupdriver=") {
			fmt.Println(strings.ReplaceAll(opt, "native.cgroupdriver=", ""))
		}
	}
}
