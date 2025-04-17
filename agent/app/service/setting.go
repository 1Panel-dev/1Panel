package service

import (
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/utils/encrypt"
	"github.com/1Panel-dev/1Panel/agent/utils/ssh"
	"github.com/jinzhu/copier"
)

type SettingService struct{}

type ISettingService interface {
	GetSettingInfo() (*dto.SettingInfo, error)
	Update(key, value string) error

	GetSSHInfo() (string, error)
	TestConnByInfo(req dto.SSHConnData) bool
	SaveConnInfo(req dto.SSHConnData) error
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

func (u *SettingService) GetSSHInfo() (string, error) {
	conn, err := settingRepo.GetValueByKey("LocalSSHConn")
	if err != nil || len(conn) == 0 {
		return "", err
	}
	return encrypt.StringDecrypt(conn)
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

	localConn, _ := json.Marshal(&connInfo)
	connAfterEncrypt, _ := encrypt.StringEncrypt(string(localConn))
	_ = settingRepo.Update("LocalSSHConn", connAfterEncrypt)
	return nil
}
