package service

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/dto/response"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/encrypt"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
	"github.com/1Panel-dev/1Panel/agent/utils/ssh"
	terminalai "github.com/1Panel-dev/1Panel/agent/utils/terminal/ai"
	"github.com/jinzhu/copier"
)

type SettingService struct{}

type ISettingService interface {
	GetSettingInfo() (*dto.SettingInfo, error)
	GetTerminalAIInfo() (*dto.TerminalAIInfo, error)
	GetFileManageAIInfo() (*dto.FileManageAIInfo, error)
	GetFileHistorySettingInfo() (*response.FileHistorySettingInfo, error)
	GetWebsiteDir() string
	Update(key, value string) error
	UpdateTerminalAI(req dto.TerminalAIInfo) error
	UpdateFileManageAI(req dto.FileManageAIInfo) error
	UpdateFileHistorySetting(req request.FileHistorySettingUpdate) error

	TestConnByInfo(req dto.SSHConnData) bool
	SaveConnInfo(req dto.SSHConnData) error
	SetDefaultIsConn(req dto.SSHDefaultConn) error
	GetSystemProxy() (*dto.SystemProxy, error)
	GetLocalConn() dto.SSHConnData
	GetLocalConnForSSH() (dto.SSHConnData, error)

	SaveDescription(req dto.CommonDescription) error
}

func NewISettingService() ISettingService {
	return &SettingService{}
}

func (u *SettingService) GetSettingInfo() (*dto.SettingInfo, error) {
	setting, err := settingRepo.GetList()
	if err != nil {
		return nil, buserr.New("ErrRecordNotFound")
	}
	settingMap := make(map[string]string)
	for _, set := range setting {
		settingMap[set.Key] = set.Value
	}
	var info dto.SettingInfo
	arr, err := json.Marshal(settingMap)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(arr, &info); err != nil {
		return nil, err
	}

	info.LocalTime = time.Now().Format("2006-01-02 15:04:05 MST -0700")
	return &info, err
}

func (u *SettingService) GetTerminalAIInfo() (*dto.TerminalAIInfo, error) {
	info := &dto.TerminalAIInfo{
		AIStatus:              constant.StatusDisable,
		AIAccountID:           "",
		AIPrefix:              constant.DefaultTerminalAIPrefix,
		AIRiskCommands:        "[]",
		AIRiskCommandsDefault: constant.DefaultTerminalAIRiskCommands,
	}

	if value, err := settingRepo.GetValueByKey("AIStatus"); err == nil && value != "" {
		info.AIStatus = value
	}
	if value, err := settingRepo.GetValueByKey("AIAccountID"); err == nil {
		info.AIAccountID = value
	}
	if value, err := settingRepo.GetValueByKey("AIPrefix"); err == nil {
		info.AIPrefix = value
	}
	if value, err := settingRepo.GetValueByKey("AIRiskCommands"); err == nil && value != "" {
		info.AIRiskCommands = value
	}

	return info, nil
}

func (u *SettingService) GetFileManageAIInfo() (*dto.FileManageAIInfo, error) {
	info := &dto.FileManageAIInfo{
		AIStatus:    constant.StatusDisable,
		AIAccountID: "",
	}
	if value, err := settingRepo.GetValueByKey("FileAIStatus"); err == nil && value != "" {
		info.AIStatus = value
	}
	if value, err := settingRepo.GetValueByKey("FileAIAccountID"); err == nil {
		info.AIAccountID = value
	}
	return info, nil
}

func (u *SettingService) GetFileHistorySettingInfo() (*response.FileHistorySettingInfo, error) {
	return historyService.GetSettingInfo()
}

func (u *SettingService) GetWebsiteDir() string {
	value, _ := settingRepo.GetValueByKey("WEBSITE_DIR")
	if value == "" {
		return path.Join(global.Dir.BaseDir, "1panel", "www")
	}
	return value
}

func (u *SettingService) Update(key, value string) error {
	oldValue := constant.FirewallPortWhiteListValue
	if key == constant.FirewallPortWhiteList {
		if _, err := firewall.ParsePortWhitelist(value); err != nil {
			return err
		}
		if val, err := settingRepo.GetValueByKey(key); err == nil {
			oldValue = val
		}
	}
	if err := settingRepo.UpdateOrCreate(key, value); err != nil {
		return err
	}
	if key == constant.FirewallPortWhiteList {
		return SyncFirewallPortWhitelistAfterUpdate(oldValue)
	}
	return nil
}

func (u *SettingService) UpdateTerminalAI(req dto.TerminalAIInfo) error {
	if strings.EqualFold(strings.TrimSpace(req.AIStatus), constant.StatusEnable) {
		accountID, err := strconv.ParseUint(strings.TrimSpace(req.AIAccountID), 10, 64)
		if err != nil || accountID == 0 {
			return buserr.New("ErrAgentAccountIDRequired")
		}
		currentStatus, _ := settingRepo.GetValueByKey("AIStatus")
		currentAccountID, _ := settingRepo.GetValueByKey("AIAccountID")
		needValidate := !strings.EqualFold(strings.TrimSpace(currentStatus), constant.StatusEnable) ||
			strings.TrimSpace(currentAccountID) != strings.TrimSpace(req.AIAccountID)
		if needValidate {
			if err := terminalai.ValidateTerminalAccount(uint(accountID)); err != nil {
				return buserr.WithErr("ErrAgentAccountUnavailable", err)
			}
		}
	}
	if err := settingRepo.UpdateOrCreate("AIStatus", req.AIStatus); err != nil {
		return err
	}
	if err := settingRepo.UpdateOrCreate("AIAccountID", req.AIAccountID); err != nil {
		return err
	}
	if err := settingRepo.UpdateOrCreate("AIPrefix", req.AIPrefix); err != nil {
		return err
	}
	if err := settingRepo.UpdateOrCreate("AIRiskCommands", req.AIRiskCommands); err != nil {
		return err
	}
	terminalai.InvalidateTerminalRuntimeCache()
	return nil
}

func (u *SettingService) UpdateFileManageAI(req dto.FileManageAIInfo) error {
	if strings.EqualFold(strings.TrimSpace(req.AIStatus), constant.StatusEnable) {
		accountID, err := strconv.ParseUint(strings.TrimSpace(req.AIAccountID), 10, 64)
		if err != nil || accountID == 0 {
			return buserr.New("ErrAgentAccountIDRequired")
		}
		currentStatus, _ := settingRepo.GetValueByKey("FileAIStatus")
		currentAccountID, _ := settingRepo.GetValueByKey("FileAIAccountID")
		needValidate := !strings.EqualFold(strings.TrimSpace(currentStatus), constant.StatusEnable) ||
			strings.TrimSpace(currentAccountID) != strings.TrimSpace(req.AIAccountID)
		if needValidate {
			if err := terminalai.ValidateTerminalAccount(uint(accountID)); err != nil {
				return buserr.WithErr("ErrAgentAccountUnavailable", err)
			}
		}
	}
	accountVal := strings.TrimSpace(req.AIAccountID)
	if !strings.EqualFold(strings.TrimSpace(req.AIStatus), constant.StatusEnable) {
		accountVal = ""
	}
	if err := settingRepo.UpdateOrCreate("FileAIStatus", req.AIStatus); err != nil {
		return err
	}
	if err := settingRepo.UpdateOrCreate("FileAIAccountID", accountVal); err != nil {
		return err
	}
	terminalai.InvalidateFileAIRuntimeCache()
	return nil
}

func (u *SettingService) UpdateFileHistorySetting(req request.FileHistorySettingUpdate) error {
	return historyService.UpdateSetting(req)
}

func (u *SettingService) TestConnByInfo(req dto.SSHConnData) bool {
	if req.AuthMode == "password" && len(req.Password) != 0 {
		password, err := base64.StdEncoding.DecodeString(req.Password)
		if err != nil {
			return false
		}
		req.Password = string(password)
	}
	if req.AuthMode == "key" && len(req.PrivateKey) != 0 {
		privateKey, err := base64.StdEncoding.DecodeString(req.PrivateKey)
		if err != nil {
			return false
		}
		req.PrivateKey = string(privateKey)
	}

	var connInfo ssh.ConnInfo
	_ = copier.Copy(&connInfo, &req)
	connInfo.PrivateKey = []byte(req.PrivateKey)
	if len(req.PassPhrase) != 0 {
		connInfo.PassPhrase = []byte(req.PassPhrase)
	}
	client, err := ssh.NewClient(connInfo)
	if err != nil {
		return false
	}
	defer client.Close()
	return true
}

func (u *SettingService) SaveConnInfo(req dto.SSHConnData) error {
	if req.AuthMode == "password" && len(req.Password) != 0 {
		password, err := base64.StdEncoding.DecodeString(req.Password)
		if err != nil {
			return err
		}
		req.Password = string(password)
	}
	if req.AuthMode == "key" && len(req.PrivateKey) != 0 {
		privateKey, err := base64.StdEncoding.DecodeString(req.PrivateKey)
		if err != nil {
			return err
		}
		req.PrivateKey = string(privateKey)
	}

	var connInfo ssh.ConnInfo
	_ = copier.Copy(&connInfo, &req)
	connInfo.PrivateKey = []byte(req.PrivateKey)
	if len(req.PassPhrase) != 0 {
		connInfo.PassPhrase = []byte(req.PassPhrase)
	}
	client, err := ssh.NewClient(connInfo)
	if err != nil {
		return err
	}
	defer client.Close()

	var connItem model.LocalConnInfo
	_ = copier.Copy(&connItem, &req)
	localConn, _ := json.Marshal(&connItem)
	connAfterEncrypt, _ := encrypt.StringEncrypt(string(localConn))
	_ = settingRepo.Update("LocalSSHConn", connAfterEncrypt)
	return nil
}

func (u *SettingService) SetDefaultIsConn(req dto.SSHDefaultConn) error {
	if req.DefaultConn == constant.StatusDisable && req.WithReset {
		if err := settingRepo.Update("LocalSSHConn", ""); err != nil {
			return err
		}
	}
	return settingRepo.Update("LocalSSHConnShow", req.DefaultConn)
}

func (u *SettingService) GetSystemProxy() (*dto.SystemProxy, error) {
	systemProxy := dto.SystemProxy{}
	systemProxy.Type, _ = settingRepo.GetValueByKey("ProxyType")
	systemProxy.URL, _ = settingRepo.GetValueByKey("ProxyUrl")
	systemProxy.Port, _ = settingRepo.GetValueByKey("ProxyPort")
	systemProxy.User, _ = settingRepo.GetValueByKey("ProxyUser")
	passwd, _ := settingRepo.GetValueByKey("ProxyPasswd")
	systemProxy.Password, _ = encrypt.StringDecrypt(passwd)
	return &systemProxy, nil
}

func (u *SettingService) loadLocalConn() dto.SSHConnData {
	var data dto.SSHConnData
	localSSHConnShow, _ := settingRepo.GetValueByKey("LocalSSHConnShow")
	data.LocalSSHConnShow = localSSHConnShow
	connItem, _ := settingRepo.GetValueByKey("LocalSSHConn")
	if len(connItem) == 0 {
		return data
	}
	connInfoInDB, _ := encrypt.StringDecrypt(connItem)
	if err := json.Unmarshal([]byte(connInfoInDB), &data); err != nil {
		data.LocalSSHConnShow = localSSHConnShow
		return data
	}
	data.LocalSSHConnShow = localSSHConnShow
	return data
}

func (u *SettingService) GetLocalConn() dto.SSHConnData {
	data := u.loadLocalConn()
	if len(data.Password) != 0 {
		data.Password = base64.StdEncoding.EncodeToString([]byte(data.Password))
	}
	if len(data.PrivateKey) != 0 {
		data.PrivateKey = base64.StdEncoding.EncodeToString([]byte(data.PrivateKey))
	}
	if len(data.PassPhrase) != 0 {
		data.PassPhrase = base64.StdEncoding.EncodeToString([]byte(data.PassPhrase))
	}
	return data
}

func (u *SettingService) GetLocalConnForSSH() (dto.SSHConnData, error) {
	data := u.loadLocalConn()
	if len(data.Addr) == 0 {
		return data, errors.New("no such ssh conn info in db")
	}
	return data, nil
}

func (u *SettingService) SaveDescription(req dto.CommonDescription) error {
	if len(req.Description) == 0 && !req.IsPinned {
		_ = settingRepo.DelDescription(req.ID)
		return nil
	}
	data, _ := settingRepo.GetDescription(settingRepo.WithByDescriptionID(req.ID), repo.WithByType(req.Type), repo.WithByDetailType(req.DetailType))
	if data.ID == "" {
		if err := copier.Copy(&data, &req); err != nil {
			return err
		}
		return settingRepo.CreateDescription(&data)
	}
	valMap := make(map[string]interface{})
	valMap["type"] = req.Type
	valMap["detail_type"] = req.DetailType
	valMap["is_pinned"] = req.IsPinned
	valMap["description"] = req.Description

	return settingRepo.UpdateDescription(data.ID, valMap)
}
