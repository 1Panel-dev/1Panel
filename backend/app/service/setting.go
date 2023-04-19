package service

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/common"
	"github.com/1Panel-dev/1Panel/backend/utils/encrypt"
	"github.com/1Panel-dev/1Panel/backend/utils/ssl"
	"github.com/gin-gonic/gin"
)

type SettingService struct{}

type ISettingService interface {
	GetSettingInfo() (*dto.SettingInfo, error)
	Update(key, value string) error
	UpdateEntrance(value string) error
	UpdatePassword(c *gin.Context, old, new string) error
	UpdatePort(port uint) error
	UpdateSSL(c *gin.Context, req dto.SSLUpdate) error
	LoadFromCert() (*dto.SSLInfo, error)
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
	if err := json.Unmarshal(arr, &info); err != nil {
		return nil, err
	}
	info.LocalTime = time.Now().Format("2006-01-02 15:04:05 MST -0700")
	return &info, err
}

func (u *SettingService) Update(key, value string) error {
	if key == "ExpirationDays" {
		timeout, _ := strconv.Atoi(value)
		if err := settingRepo.Update("ExpirationTime", time.Now().AddDate(0, 0, timeout).Format("2006-01-02 15:04:05")); err != nil {
			return err
		}
	}
	if key == "SecurityEntrance" {
		if err := settingRepo.Update("SecurityEntranceStatus", "enable"); err != nil {
			return err
		}
	}
	if key == "SecurityEntranceStatus" {
		if err := settingRepo.Update("SecurityEntrance", ""); err != nil {
			return err
		}
	}
	if err := settingRepo.Update(key, value); err != nil {
		return err
	}
	if key == "UserName" {
		_ = global.SESSION.Clean()
	}
	return nil
}

func (u *SettingService) UpdateEntrance(value string) error {
	if err := settingRepo.Update("SecurityEntranceStatus", "enable"); err != nil {
		return err
	}
	if err := settingRepo.Update("SecurityEntrance", value); err != nil {
		return err
	}
	return nil
}

func (u *SettingService) UpdatePort(port uint) error {
	if common.ScanPort(int(port)) {
		return buserr.WithDetail(constant.ErrPortInUsed, port, nil)
	}
	serverPort, err := settingRepo.Get(settingRepo.WithByKey("ServerPort"))
	if err != nil {
		return err
	}
	portValue, _ := strconv.Atoi(serverPort.Value)
	if err := OperateFirewallPort([]int{portValue}, []int{int(port)}); err != nil {
		global.LOG.Errorf("set system firewall ports failed, err: %v", err)
	}
	if err := settingRepo.Update("ServerPort", strconv.Itoa(int(port))); err != nil {
		return err
	}
	go func() {
		_, err := cmd.Exec("systemctl restart 1panel.service")
		if err != nil {
			global.LOG.Errorf("restart system port failed, err: %v", err)
		}
	}()
	return nil
}

func (u *SettingService) UpdateSSL(c *gin.Context, req dto.SSLUpdate) error {
	if req.SSL == "disable" {
		if err := settingRepo.Update("SSL", "disable"); err != nil {
			return err
		}
		if err := settingRepo.Update("SSLType", "self"); err != nil {
			return err
		}
		_ = os.Remove(fmt.Sprintf("%s/1panel/secret/cert.pem", global.CONF.System.BaseDir))
		_ = os.Remove(fmt.Sprintf("%s/1panel/secret/key.pem", global.CONF.System.BaseDir))
		go func() {
			_, err := cmd.Exec("systemctl restart 1panel.service")
			if err != nil {
				global.LOG.Errorf("restart system failed, err: %v", err)
			}
		}()
		return nil
	}

	switch req.SSLType {
	case "self":
		domains := loadDomain(c)
		if err := ssl.GenerateSSL(domains); err != nil {
			return err
		}
	case "import":
		cert, err := os.OpenFile("/opt/1panel/secret/cert.pem", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}
		defer cert.Close()
		if _, err := cert.WriteString(req.Cert); err != nil {
			return err
		}
		key, err := os.OpenFile("/opt/1panel/secret/cert.pem", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}
		if _, err := key.WriteString(req.Key); err != nil {
			return err
		}
		defer key.Close()
	}
	if err := settingRepo.Update("SSL", req.SSL); err != nil {
		return err
	}
	if err := settingRepo.Update("SSLType", req.SSLType); err != nil {
		return err
	}
	go func() {
		_, err := cmd.Exec("systemctl restart 1panel.service")
		if err != nil {
			global.LOG.Errorf("restart system failed, err: %v", err)
		}
	}()
	return nil
}

func (u *SettingService) LoadFromCert() (*dto.SSLInfo, error) {
	certFile := global.CONF.System.BaseDir + "/1panel/secret/user.crt"
	certData, err := os.ReadFile(certFile)
	if err != nil {
		return nil, err
	}
	certBlock, _ := pem.Decode(certData)
	if certBlock == nil {
		return nil, err
	}
	certObj, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, err
	}
	return &dto.SSLInfo{
		Domain:   strings.Join(certObj.DNSNames, ","),
		Subject:  certObj.Subject.CommonName,
		Timeout:  certObj.NotAfter.Format("2006-01-02 15:04:05"),
		RootPath: global.CONF.System.BaseDir + "/1panel/secret/root.crt",
	}, nil
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
		if err := settingRepo.Update("ExpirationTime", time.Now().AddDate(0, 0, timeout).Format("2006-01-02 15:04:05")); err != nil {
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
	_ = global.SESSION.Clean()
	return nil
}

func loadDomain(c *gin.Context) []string {
	var domain []string
	ip := c.Request.RemoteAddr
	if idx := strings.Index(ip, ":"); idx != -1 {
		ip = ip[:idx]
		if ip != "localhost" && ip != "127.0.0.1" && ip != "::1" {
			domain = append(domain, ip)
		}
	}
	if host := c.Request.Header.Get("X-Forwarded-Host"); len(host) > 0 {
		domain = append(domain, host)
	} else if host := c.Request.Header.Get("Host"); len(host) > 0 {
		domain = append(domain, host)
	}
	return domain
}
