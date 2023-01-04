package job

import (
	"github.com/1Panel-dev/1Panel/backend/app/repo"
	"github.com/1Panel-dev/1Panel/backend/app/service"
	"github.com/1Panel-dev/1Panel/backend/global"
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
	global.LOG.Info("ssl renew cron job start...")
	now := time.Now()
	for _, s := range sslList {
		if !s.AutoRenew || s.Provider == "manual" || s.Provider == "dnsManual" {
			continue
		}
		sum := s.ExpireDate.Sub(now)
		if sum.Hours() < 168 {
			if err := sslService.Renew(s.ID); err != nil {
				global.LOG.Errorf("renew doamin [%s] ssl failed err:%s", s.PrimaryDomain, err.Error())
			}
		}
	}

	global.LOG.Info("ssl renew cron job end...")
}
