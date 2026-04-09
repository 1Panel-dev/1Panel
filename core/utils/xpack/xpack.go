//go:build xpack

package xpack

import (
	"errors"
	"net/http"

	baseDto "github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/init/session/psession"
	"github.com/1Panel-dev/1Panel/core/utils/ssh"
	"github.com/1Panel-dev/1Panel/core/xpack/app/model"
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

func Login(_ *gin.Context, _ baseDto.Login, _ string) (*baseDto.UserLoginInfo, string, error) {
	return nil, "", errors.New("not xpackee build")
}

func LoadSessionTimeout(sessionUser psession.SessionUser, defaultTTL int) int { return defaultTTL }

func RemoveBindNode(nodeID uint) error { return nil }

func CheckLicenseStatus(isXpack bool, licenseID, nodeID uint) error {
	return edition.CheckLicenseStatus(isXpack, licenseID, nodeID)
}
func LoadLicenseByNodeID(nodeID uint) (bool, model.License, error) {
	return edition.LoadLicenseByNodeID(nodeID)
}
func Bind(node model.Node, license *model.License, withInit, withSync, withDryRun, withDockerRestart bool) error {
	return edition.Bind(node, license, withInit, withSync, withDryRun, withDockerRestart)
}
func BindFree(node model.Node, licenseID uint) error {
	return edition.BindFree(node, licenseID)
}
func Unbind(node model.Node, license model.License, withReset, withSync, withDockerRestart bool) error {
	return edition.Unbind(node, license, withReset, withSync, withDockerRestart)
}
func UnbindFree(node model.Node, licenseID uint) error {
	return edition.UnbindFree(node, licenseID)
}
