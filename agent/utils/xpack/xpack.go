//go:build !xpack

package xpack

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
)

func RemoveTamper(website string) {}

func LoadRequestTransport() *http.Transport {
	return &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: (&net.Dialer{
			Timeout:   60 * time.Second,
			KeepAlive: 60 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		IdleConnTimeout:       15 * time.Second,
	}
}

func LoadGpuInfo() []interface{} {
	return nil
}

func StartClam(startClam model.Clam, isUpdate bool) (int, error) {
	return 0, buserr.New(constant.ErrXpackNotFound)
}

func LoadNodeInfo() (bool, model.NodeInfo, error) {
	var info model.NodeInfo
	info.BaseDir = loadParams("BASE_DIR")
	info.Version = loadParams("ORIGINAL_VERSION")
	info.CurrentNode = "127.0.0.1"
	info.EncryptKey = common.RandStr(16)
	return false, info, nil
}

func loadParams(param string) string {
	stdout, err := cmd.Execf("grep '^%s=' /usr/local/bin/1pctl | cut -d'=' -f2", param)
	if err != nil {
		panic(err)
	}
	info := strings.ReplaceAll(stdout, "\n", "")
	if len(info) == 0 || info == `""` {
		panic(fmt.Sprintf("error `%s` find in /usr/local/bin/1pctl", param))
	}
	return info
}

func RequestToMaster(reqUrl, reqMethod string, reqBody io.Reader) (interface{}, error) {
	return nil, nil
}

func GetImagePrefix() string {
	return ""
}
