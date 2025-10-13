package service

import (
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/utils/encrypt"
	"github.com/1Panel-dev/1Panel/agent/utils/ssh"
	"github.com/jinzhu/copier"
)

type SettingService struct{}

type ISettingService interface {
	GetSettingInfo() (*dto.SettingInfo, error)
	Update(key, value string) error

	TestConnByInfo(req dto.SSHConnData) bool
	SaveConnInfo(req dto.SSHConnData) error
	GetSystemProxy() (*dto.SystemProxy, error)
	GetLocalConn() dto.SSHConnData
	GetSettingByKey(key string) string
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

func (u *SettingService) Update(key, value string) error {
	return settingRepo.UpdateOrCreate(key, value)
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

func (u *SettingService) GetLocalConn() dto.SSHConnData {
	var data dto.SSHConnData
	connItem, _ := settingRepo.GetValueByKey("LocalSSHConn")
	if len(connItem) == 0 {
		return data
	}
	connInfoInDB, _ := encrypt.StringDecrypt(connItem)
	data.LocalSSHConnShow, _ = settingRepo.GetValueByKey("LocalSSHConnShow")
	if err := json.Unmarshal([]byte(connInfoInDB), &data); err != nil {
		return data
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
	return data
}

func (u *SettingService) GetSettingByKey(key string) string {
	switch key {
	case "LocalSSHConn":
		value, _ := settingRepo.GetValueByKey(key)
		if len(value) == 0 {
			return ""
		}
		itemStr, _ := encrypt.StringDecrypt(value)
		return itemStr
	default:
		value, _ := settingRepo.GetValueByKey(key)
		return value
	}
}
