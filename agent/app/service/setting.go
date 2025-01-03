package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/encrypt"
	"github.com/1Panel-dev/1Panel/agent/utils/xpack"
)

type SettingService struct{}

type ISettingService interface {
	GetSettingInfo() (*dto.SettingInfo, error)
	Update(key, value string) error

	ReloadConn() error
}

func NewISettingService() ISettingService {
	return &SettingService{}
}

func (u *SettingService) GetSettingInfo() (*dto.SettingInfo, error) {
	setting, err := settingRepo.GetList()
	if err != nil {
		return nil, constant.ErrRecordNotFound
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

func (u *SettingService) ReloadConn() error {
	if global.IsMaster {
		return nil
	}
	isLocal, nodeInfo, err := xpack.LoadNodeInfo()
	if err != nil {
		global.LOG.Errorf("load new node info failed, err: %v", err)
		return nil
	}
	if isLocal {
		global.IsMaster = true
		return nil
	}

	itemKey, _ := encrypt.StringEncrypt(nodeInfo.ServerKey)
	if err := settingRepo.Update("ServerKey", itemKey); err != nil {
		global.LOG.Errorf("update server key failed, err: %v", err)
		return nil
	}
	itemCrt, _ := encrypt.StringEncrypt(nodeInfo.ServerCrt)
	if err := settingRepo.Update("ServerCrt", itemCrt); err != nil {
		global.LOG.Errorf("update server crt failed, err: %v", err)
		return nil
	}
	if err := settingRepo.Update("NodeScope", nodeInfo.Scope); err != nil {
		global.LOG.Errorf("update current node failed, err: %v", err)
		return nil
	}
	if err := settingRepo.Update("SystemVersion", nodeInfo.Version); err != nil {
		global.LOG.Errorf("update system version failed, err: %v", err)
		return nil
	}
	if err := settingRepo.Update("BaseDir", nodeInfo.BaseDir); err != nil {
		global.LOG.Errorf("update base dir failed, err: %v", err)
		return nil
	}
	if err := settingRepo.Update("NodePort", fmt.Sprintf("%v", nodeInfo.NodePort)); err != nil {
		global.LOG.Errorf("update node port failed, err: %v", err)
		return nil
	}

	global.CONF.System.BaseDir = nodeInfo.BaseDir
	global.CONF.System.Version = nodeInfo.Version
	global.CONF.System.Port = fmt.Sprintf("%v", nodeInfo.NodePort)
	if global.CONF.System.Port == "0" {
		global.CONF.System.Port = "9999"
	}
	global.IsMaster = nodeInfo.Scope == "master"
	return nil
}
