//go:build xpack

package xpack

import (
	"net/http"

	"github.com/1Panel-dev/1Panel/core/utils/ssh"
	edition "github.com/1Panel-dev/1Panel/core/xpack/edition"
	"github.com/gin-gonic/gin"
)

func Proxy(c *gin.Context, currentNode string) {
	edition.Proxy(c, currentNode)
}

func ProxyDocker(proxyURL string) error { return edition.ProxyDocker(proxyURL) }

func UpdateGroup(name string, group, newGroup uint) error {
	return edition.UpdateGroup(name, group, newGroup)
}

func CheckBackupUsed(name string) error {
	return edition.CheckBackupUsed(name)
}

func LoadRequestTransport() *http.Transport { return edition.LoadRequestTransport() }

func LoadNodeInfo(currentNode string) (*ssh.ConnInfo, string, error) {
	return edition.LoadNodeInfo(currentNode)
}

func Sync(dataType string) error { return edition.Sync(dataType) }

func AutoUpgradeWithMaster() { edition.AutoUpgradeWithMaster() }
