package job

import (
	"path"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/app/service"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
)

type ssl struct {
}

func NewSSLJob() *ssl {
	return &ssl{}
}

func (ssl *ssl) Run() {
	systemSSLEnable, sslID := service.GetSystemSSL()
	sslRepo := repo.NewISSLRepo()
	sslService := service.NewIWebsiteSSLService()
	sslList, _ := sslRepo.List()
	nyc, _ := time.LoadLocation(common.LoadTimeZone())
	global.LOG.Info("The scheduled certificate update task is currently in progress ...")
	now := time.Now().Add(10 * time.Second)
	for _, s := range sslList {
		if !s.AutoRenew || s.Provider == "manual" || s.Provider == "dnsManual" || s.Status == "applying" {
			continue
		}
		expireDate := s.ExpireDate.In(nyc)
		sub := expireDate.Sub(now)
		if sub.Hours() < 720 {
			global.LOG.Infof("Update the SSL certificate for the [%s] domain", s.PrimaryDomain)
			if s.Provider == constant.SelfSigned {
				caService := service.NewIWebsiteCAService()
				if _, err := caService.ObtainSSL(request.WebsiteCAObtain{
					ID:    s.CaID,
					SSLID: s.ID,
					Renew: true,
					Unit:  "year",
					Time:  10,
				}); err != nil {
					global.LOG.Errorf("Failed to update the SSL certificate for the [%s] domain , err:%s", s.PrimaryDomain, err.Error())
					continue
				}
			} else {
				if err := sslService.ObtainSSL(request.WebsiteSSLApply{
					ID: s.ID,
				}); err != nil {
					global.LOG.Errorf("Failed to update the SSL certificate for the [%s] domain , err:%s", s.PrimaryDomain, err.Error())
					continue
				}
			}
			if systemSSLEnable && sslID == s.ID {
				websiteSSL, _ := sslRepo.GetFirst(repo.NewCommonRepo().WithByID(s.ID))
				fileOp := files.NewFileOp()
				secretDir := path.Join(global.CONF.System.BaseDir, "1panel/secret")
				if err := fileOp.WriteFile(path.Join(secretDir, "server.crt"), strings.NewReader(websiteSSL.Pem), 0600); err != nil {
					global.LOG.Errorf("Failed to update the SSL certificate File for 1Panel System domain [%s] , err:%s", s.PrimaryDomain, err.Error())
					continue
				}
				if err := fileOp.WriteFile(path.Join(secretDir, "server.key"), strings.NewReader(websiteSSL.PrivateKey), 0600); err != nil {
					global.LOG.Errorf("Failed to update the SSL certificate for 1Panel System domain [%s] , err:%s", s.PrimaryDomain, err.Error())
					continue
				}
			}
			global.LOG.Infof("The SSL certificate for the [%s] domain has been successfully updated", s.PrimaryDomain)
		}
	}
	global.LOG.Info("The scheduled certificate update task has completed")
}
