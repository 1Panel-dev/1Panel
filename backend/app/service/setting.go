package service

import (
	"crypto/tls"
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
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"github.com/1Panel-dev/1Panel/backend/utils/ntp"
	"github.com/1Panel-dev/1Panel/backend/utils/ssl"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
)

type SettingService struct{}

type ISettingService interface {
	GetSettingInfo() (*dto.SettingInfo, error)
	LoadTimeZone() ([]string, error)
	Update(key, value string) error
	UpdatePassword(c *gin.Context, old, new string) error
	UpdatePort(port uint) error
	UpdateSSL(c *gin.Context, req dto.SSLUpdate) error
	LoadFromCert() (*dto.SSLInfo, error)
	HandlePasswordExpired(c *gin.Context, old, new string) error
	SyncTime(req dto.SyncTime) error
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

func (u *SettingService) LoadTimeZone() ([]string, error) {
	std, err := cmd.Exec("timedatectl list-timezones")
	if err != nil {
		return []string{}, nil
	}
	return strings.Split(std, "\n"), err
}

func (u *SettingService) Update(key, value string) error {
	switch key {
	case "MonitorStatus":
		if value == "enable" && global.MonitorCronID == 0 {
			interval, err := settingRepo.Get(settingRepo.WithByKey("MonitorInterval"))
			if err != nil {
				return err
			}
			if err := StartMonitor(false, interval.Value); err != nil {
				return err
			}
		}
		if value == "disable" && global.MonitorCronID != 0 {
			global.Cron.Remove(cron.EntryID(global.MonitorCronID))
			global.MonitorCronID = 0
		}
	case "MonitorInterval":
		status, err := settingRepo.Get(settingRepo.WithByKey("MonitorStatus"))
		if err != nil {
			return err
		}
		if status.Value == "enable" && global.MonitorCronID != 0 {
			if err := StartMonitor(true, value); err != nil {
				return err
			}
		}
	case "TimeZone":
		if err := ntp.UpdateSystemTimeZone(value); err != nil {
			return err
		}
	}

	if err := settingRepo.Update(key, value); err != nil {
		return err
	}

	switch key {
	case "ExpirationDays":
		timeout, _ := strconv.Atoi(value)
		if err := settingRepo.Update("ExpirationTime", time.Now().AddDate(0, 0, timeout).Format("2006-01-02 15:04:05")); err != nil {
			return err
		}
	case "TimeZone":
		go func() {
			_, err := cmd.Exec("systemctl restart 1panel.service")
			if err != nil {
				global.LOG.Errorf("restart system for new time zone failed, err: %v", err)
			}
		}()
	case "BindDomain":
		if len(value) != 0 {
			_ = global.SESSION.Clean()
		}
	case "UserName", "Password":
		_ = global.SESSION.Clean()
	}

	return nil
}

func (u *SettingService) SyncTime(req dto.SyncTime) error {
	if err := settingRepo.Update("NtpSite", req.NtpSite); err != nil {
		return err
	}
	ntime, err := ntp.GetRemoteTime(req.NtpSite)
	if err != nil {
		return err
	}
	ts := ntime.Format("2006-01-02 15:04:05")
	if err := ntp.UpdateSystemTime(ts); err != nil {
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
	secretDir := global.CONF.System.BaseDir + "/1panel/secret/"
	if req.SSL == "disable" {
		if err := settingRepo.Update("SSL", "disable"); err != nil {
			return err
		}
		if err := settingRepo.Update("SSLType", "self"); err != nil {
			return err
		}
		_ = os.Remove(secretDir + "server.crt")
		_ = os.Remove(secretDir + "server.key")
		go func() {
			_, err := cmd.Exec("systemctl restart 1panel.service")
			if err != nil {
				global.LOG.Errorf("restart system failed, err: %v", err)
			}
		}()
		return nil
	}

	if _, err := os.Stat(secretDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(secretDir, os.ModePerm); err != nil {
			return err
		}
	}
	if err := settingRepo.Update("SSLType", req.SSLType); err != nil {
		return err
	}
	if req.SSLType == "self" {
		if len(req.Domain) == 0 {
			return fmt.Errorf("load domain failed")
		}
		if err := ssl.GenerateSSL(req.Domain); err != nil {
			return err
		}
	}
	if req.SSLType == "select" {
		sslInfo, err := websiteSSLRepo.GetFirst(commonRepo.WithByID(req.SSLID))
		if err != nil {
			return err
		}
		req.Cert = sslInfo.Pem
		req.Key = sslInfo.PrivateKey
		req.SSLType = "import"
		if err := settingRepo.Update("SSLID", strconv.Itoa(int(req.SSLID))); err != nil {
			return err
		}
	}
	if req.SSLType == "import" {
		cert, err := os.OpenFile(secretDir+"server.crt.tmp", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}
		defer cert.Close()
		if _, err := cert.WriteString(req.Cert); err != nil {
			return err
		}
		key, err := os.OpenFile(secretDir+"server.key.tmp", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}
		if _, err := key.WriteString(req.Key); err != nil {
			return err
		}
		defer key.Close()
	}
	if err := checkCertValid(req.Domain); err != nil {
		return err
	}

	fileOp := files.NewFileOp()
	if err := fileOp.Rename(secretDir+"server.crt.tmp", secretDir+"server.crt"); err != nil {
		return err
	}
	if err := fileOp.Rename(secretDir+"server.key.tmp", secretDir+"server.key"); err != nil {
		return err
	}
	if err := settingRepo.Update("SSL", req.SSL); err != nil {
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
	ssl, err := settingRepo.Get(settingRepo.WithByKey("SSL"))
	if err != nil {
		return nil, err
	}
	if ssl.Value == "disable" {
		return &dto.SSLInfo{}, nil
	}
	sslType, err := settingRepo.Get(settingRepo.WithByKey("SSLType"))
	if err != nil {
		return nil, err
	}
	data, err := loadInfoFromCert()
	if err != nil {
		return nil, err
	}
	switch sslType.Value {
	case "import":
		if _, err := os.Stat(global.CONF.System.BaseDir + "/1panel/secret/server.crt"); err != nil {
			return nil, fmt.Errorf("load server.crt file failed, err: %v", err)
		}
		certFile, _ := os.ReadFile(global.CONF.System.BaseDir + "/1panel/secret/server.crt")
		data.Cert = string(certFile)

		if _, err := os.Stat(global.CONF.System.BaseDir + "/1panel/secret/server.key"); err != nil {
			return nil, fmt.Errorf("load server.key file failed, err: %v", err)
		}
		keyFile, _ := os.ReadFile(global.CONF.System.BaseDir + "/1panel/secret/server.key")
		data.Key = string(keyFile)
	case "select":
		sslID, err := settingRepo.Get(settingRepo.WithByKey("SSLID"))
		if err != nil {
			return nil, err
		}
		id, _ := strconv.Atoi(sslID.Value)
		data.SSLID = uint(id)
	}
	return data, nil
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

func loadInfoFromCert() (*dto.SSLInfo, error) {
	var info dto.SSLInfo
	certFile := global.CONF.System.BaseDir + "/1panel/secret/server.crt"
	if _, err := os.Stat(certFile); err != nil {
		return &info, err
	}
	certData, err := os.ReadFile(certFile)
	if err != nil {
		return &info, err
	}
	certBlock, _ := pem.Decode(certData)
	if certBlock == nil {
		return &info, err
	}
	certObj, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return &info, err
	}
	var domains []string
	if len(certObj.IPAddresses) != 0 {
		for _, ip := range certObj.IPAddresses {
			domains = append(domains, ip.String())
		}
	}
	if len(certObj.DNSNames) != 0 {
		domains = append(domains, certObj.DNSNames...)
	}
	return &dto.SSLInfo{
		Domain:   strings.Join(domains, ","),
		Timeout:  certObj.NotAfter.Format("2006-01-02 15:04:05"),
		RootPath: global.CONF.System.BaseDir + "/1panel/secret/server.crt",
	}, nil
}

func checkCertValid(domain string) error {
	certificate, err := os.ReadFile(global.CONF.System.BaseDir + "/1panel/secret/server.crt.tmp")
	if err != nil {
		return err
	}
	key, err := os.ReadFile(global.CONF.System.BaseDir + "/1panel/secret/server.key.tmp")
	if err != nil {
		return err
	}
	if _, err = tls.X509KeyPair(certificate, key); err != nil {
		return err
	}
	certBlock, _ := pem.Decode(certificate)
	if certBlock == nil {
		return err
	}
	if _, err := x509.ParseCertificate(certBlock.Bytes); err != nil {
		return err
	}

	return nil
}
