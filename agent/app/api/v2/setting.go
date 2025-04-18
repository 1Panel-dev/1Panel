package v2

import (
	"encoding/json"
	"os"
	"os/user"
	"path"

	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/ssh"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// @Tags System Setting
// @Summary Load system setting info
// @Success 200 {object} dto.SettingInfo
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /settings/search [post]
func (b *BaseApi) GetSettingInfo(c *gin.Context) {
	setting, err := settingService.GetSettingInfo()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, setting)
}

// @Tags System Setting
// @Summary Load system available status
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /settings/search/available [get]
func (b *BaseApi) GetSystemAvailable(c *gin.Context) {
	helper.Success(c)
}

// @Tags System Setting
// @Summary Update system setting
// @Accept json
// @Param request body dto.SettingUpdate true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /settings/update [post]
// @x-panel-log {"bodyKeys":["key","value"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"修改系统配置 [key] => [value]","formatEN":"update system setting [key] => [value]"}
func (b *BaseApi) UpdateSetting(c *gin.Context) {
	var req dto.SettingUpdate
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := settingService.Update(req.Key, req.Value); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

// @Tags System Setting
// @Summary Load local backup dir
// @Success 200 {string} path
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /settings/basedir [get]
func (b *BaseApi) LoadBaseDir(c *gin.Context) {
	helper.SuccessWithData(c, global.Dir.DataDir)
}

func (b *BaseApi) CheckLocalConn(c *gin.Context) {
	_, err := loadLocalConn()
	helper.SuccessWithData(c, err == nil)
}

// @Tags System Setting
// @Summary Check local conn info
// @Success 200 {bool} isOk
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /settings/ssh/check/info [post]
func (b *BaseApi) CheckLocalConnByInfo(c *gin.Context) {
	var req dto.SSHConnData
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	helper.SuccessWithData(c, settingService.TestConnByInfo(req))
}

// @Tags System Setting
// @Summary Save local conn info
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /settings/ssh [post]
func (b *BaseApi) SaveLocalConnInfo(c *gin.Context) {
	var req dto.SSHConnData
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	helper.SuccessWithData(c, settingService.SaveConnInfo(req))
}

func loadLocalConn() (*ssh.SSHClient, error) {
	itemPath := ""
	currentInfo, _ := user.Current()
	if len(currentInfo.HomeDir) == 0 {
		itemPath = "/root/.ssh/id_ed25519_1panel"
	} else {
		itemPath = path.Join(currentInfo.HomeDir, ".ssh/id_ed25519_1panel")
	}
	if _, err := os.Stat(itemPath); err != nil {
		_ = sshService.GenerateSSH(dto.GenerateSSH{EncryptionMode: "ed25519", Name: "_1panel"})
	}

	privateKey, _ := os.ReadFile(itemPath)
	connWithKey := ssh.ConnInfo{
		Addr:       "127.0.0.1",
		User:       "root",
		Port:       22,
		AuthMode:   "key",
		PrivateKey: privateKey,
	}
	client, err := ssh.NewClient(connWithKey)
	if err == nil {
		return client, nil
	}

	connInfoInDB, err := settingService.GetSSHInfo()
	if err != nil {
		return nil, err
	}
	if len(connInfoInDB) == 0 {
		return nil, errors.New("no such ssh conn info in db!")
	}
	var connInDB ssh.ConnInfo
	if err := json.Unmarshal([]byte(connInfoInDB), &connInDB); err != nil {
		return nil, err
	}
	return ssh.NewClient(connInDB)
}
