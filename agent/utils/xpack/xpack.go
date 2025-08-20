//go:build !xpack

package xpack

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/gin-gonic/gin"
)

func RemoveTamper(website string) {}

func StartClam(startClam model.Clam, isUpdate bool) (int, error) {
	return 0, buserr.New("ErrXpackNotFound")
}

func LoadNodeInfo(isBase bool) (model.NodeInfo, error) {
	var info model.NodeInfo
	info.BaseDir = loadParams("BASE_DIR")
	info.Version = loadParams("ORIGINAL_VERSION")
	info.Scope = "master"
	global.IsMaster = true
	return info, nil
}

func loadParams(param string) string {
	stdout, err := cmd.RunDefaultWithStdoutBashCf("grep '^%s=' /usr/local/bin/1pctl | cut -d'=' -f2", param)
	if err != nil {
		panic(err)
	}
	info := strings.ReplaceAll(stdout, "\n", "")
	if len(info) == 0 || info == `""` {
		panic(fmt.Sprintf("error `%s` find in /usr/local/bin/1pctl", param))
	}
	return info
}

func GetImagePrefix() string {
	return ""
}

func IsUseCustomApp() bool {
	return false
}

func CreateTaskScanSMSAlertLog(alert dto.AlertDTO, alertType string, create dto.AlertLogCreate, pushAlert dto.PushAlert, method string) error {
	return nil
}

func CreateSMSAlertLog(alertType string, info dto.AlertDTO, create dto.AlertLogCreate, project string, params []dto.Param, method string) error {
	return nil
}

func GetLicenseErrorAlert() (uint, error) {
	return 0, nil
}

func GetNodeErrorAlert() (uint, error) {
	return 0, nil
}

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

func ValidateCertificate(c *gin.Context) bool {
	return true
}

func PushSSLToNode(websiteSSL *model.WebsiteSSL) error {
	return nil
}
