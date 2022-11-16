package service

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/encrypt"
	"github.com/gin-gonic/gin"
)

type SettingService struct{}

type ISettingService interface {
	GetSettingInfo() (*dto.SettingInfo, error)
	Update(c *gin.Context, key, value string) error
	UpdatePassword(c *gin.Context, old, new string) error
	HandlePasswordExpired(c *gin.Context, old, new string) error
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
	_ = json.Unmarshal(arr, &info)
	info.MonitorStoreDays, _ = strconv.Atoi(settingMap["MonitorStoreDays"])
	info.ServerPort, _ = strconv.Atoi(settingMap["ServerPort"])
	info.SessionTimeout, _ = strconv.Atoi(settingMap["SessionTimeout"])
	info.ExpirationDays, _ = strconv.Atoi(settingMap["ExpirationDays"])
	info.LocalTime = time.Now().Format("2006-01-02 15:04:05")
	return &info, err
}

func (u *SettingService) Update(c *gin.Context, key, value string) error {
	if key == "ExpirationDays" {
		timeout, _ := strconv.Atoi(value)
		if err := settingRepo.Update("ExpirationTime", time.Now().AddDate(0, 0, timeout).Format("2006.01.02 15:04:05")); err != nil {
			return err
		}
		c.SetCookie(constant.PasswordExpiredName, encrypt.Md5(time.Now().Format("20060102150405")), 86400*timeout, "", "", false, false)
	}
	if err := settingRepo.Update(key, value); err != nil {
		return err
	}
	return nil
}

func (u *SettingService) HandlePasswordExpired(c *gin.Context, old, new string) error {
	setting, err := settingRepo.Get(settingRepo.WithByKey("Password"))
	if err != nil {
		return err
	}
	passwordFromDB, err := encrypt.StringDecrypt(setting.Value)
	if err != nil {
		return err
	}
	if passwordFromDB == old {
		newPassword, err := encrypt.StringEncrypt(new)
		if err != nil {
			return err
		}
		if err := settingRepo.Update("Password", newPassword); err != nil {
			return err
		}

		expiredSetting, err := settingRepo.Get(settingRepo.WithByKey("ExpirationDays"))
		if err != nil {
			return err
		}
		timeout, _ := strconv.Atoi(expiredSetting.Value)
		c.SetCookie(constant.PasswordExpiredName, encrypt.Md5(time.Now().Format("20060102150405")), 86400*timeout, "", "", false, false)
		if err := settingRepo.Update("ExpirationTime", time.Now().AddDate(0, 0, timeout).Format("2006.01.02 15:04:05")); err != nil {
			return err
		}
		return nil
	}
	return constant.ErrInitialPassword
}

func (u *SettingService) UpdatePassword(c *gin.Context, old, new string) error {
	if err := u.HandlePasswordExpired(c, old, new); err != nil {
		return err
	}
	sID, _ := c.Cookie(constant.SessionName)
	if sID != "" {
		c.SetCookie(constant.SessionName, sID, -1, "", "", false, false)
		err := global.SESSION.Delete(sID)
		if err != nil {
			return err
		}
	}
	return nil
}
