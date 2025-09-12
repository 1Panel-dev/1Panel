package v2

import (
	"encoding/base64"
	"encoding/json"

	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
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

// @Tags System Setting
// @Summary Load local conn
// @Success 200 {object} dto.SSHConnData
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /settings/ssh/conn [get]
func (b *BaseApi) LoadLocalConn(c *gin.Context) {
	connInfoInDB := settingService.GetSettingByKey("LocalSSHConn")
	if len(connInfoInDB) == 0 {
		helper.Success(c)
		return
	}
	var data dto.SSHConnData
	if err := json.Unmarshal([]byte(connInfoInDB), &data); err != nil {
		helper.Success(c)
		return
	}
	if len(data.Password) != 0 {
		data.Password = base64.StdEncoding.EncodeToString([]byte(data.Password))
	}
	if len(data.PrivateKey) != 0 {
		data.PrivateKey = base64.StdEncoding.EncodeToString([]byte(data.PrivateKey))
	}
	if len(data.PassPhrase) != 0 {
		data.PassPhrase = base64.StdEncoding.EncodeToString([]byte(data.PassPhrase))
	}
	helper.SuccessWithData(c, data)
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
func (b *BaseApi) SaveLocalConn(c *gin.Context) {
	var req dto.SSHConnData
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	if err := settingService.SaveConnInfo(req); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}

func loadLocalConn() (*ssh.SSHClient, error) {
	connInfoInDB := settingService.GetSettingByKey("LocalSSHConn")
	if len(connInfoInDB) == 0 {
		return nil, errors.New("no such ssh conn info in db!")
	}
	var connInDB model.LocalConnInfo
	if err := json.Unmarshal([]byte(connInfoInDB), &connInDB); err != nil {
		return nil, err
	}
	sshInfo := ssh.ConnInfo{
		Addr:       connInDB.Addr,
		Port:       int(connInDB.Port),
		User:       connInDB.User,
		AuthMode:   connInDB.AuthMode,
		Password:   connInDB.Password,
		PrivateKey: []byte(connInDB.PrivateKey),
		PassPhrase: []byte(connInDB.PassPhrase),
	}
	return ssh.NewClient(sshInfo)
}

// @Tags System Setting
// @Summary Load system setting by key
// @Param key path string true "key"
// @Success 200 {object} dto.SettingInfo
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /settings/get/{key} [get]
func (b *BaseApi) GetSettingByKey(c *gin.Context) {
	key := c.Param("key")
	if len(key) == 0 {
		helper.BadRequest(c, errors.New("key is empty"))
		return
	}
	value := settingService.GetSettingByKey(key)
	helper.SuccessWithData(c, value)
}
