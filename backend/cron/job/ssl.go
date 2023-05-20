package job

import (
	"github.com/1Panel-dev/1Panel/backend/app/repo"
	"github.com/1Panel-dev/1Panel/backend/app/service"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/common"
	"time"
)

type ssl struct {
}

func NewSSLJob() *ssl {
	return &ssl{}
}

func (ssl *ssl) Run() {
	sslRepo := repo.NewISSLRepo()
	sslService := service.NewIWebsiteSSLService()
	sslList, _ := sslRepo.List()
	nyc, _ := time.LoadLocation(common.LoadTimeZone())
	global.LOG.Info("The scheduled certificate update task is currently in progress ...")
	now := time.Now().Add(10 * time.Second)
	for _, s := range sslList {
		if !s.AutoRenew || s.Provider == "manual" || s.Provider == "dnsManual" {
			continue
		}
		expireDate := s.ExpireDate.In(nyc)
		sub := expireDate.Sub(now)
		if sub.Hours() < 720 {
			global.LOG.Errorf("Update the SSL certificate for the [%s] domain", s.PrimaryDomain)
			if err := sslService.Renew(s.ID); err != nil {
				global.LOG.Errorf("Failed to update the SSL certificate for the [%s] domain , err:%s", s.PrimaryDomain, err.Error())
				continue
			}
			global.LOG.Errorf("The SSL certificate for the [%s] domain has been successfully updated", s.PrimaryDomain)
		}
	}
	global.LOG.Info("The scheduled certificate update task has completed")
}
