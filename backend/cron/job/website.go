package job

import (
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/app/repo"
	"github.com/1Panel-dev/1Panel/backend/app/service"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/common"
)

type website struct{}

func NewWebsiteJob() *website {
	return &website{}
}

func (w *website) Run() {
	nyc, _ := time.LoadLocation(common.LoadTimeZoneByCmd())
	websites, _ := repo.NewIWebsiteRepo().List()
	global.LOG.Info("Website scheduled task in progress ...")
	now := time.Now().Add(10 * time.Minute)
	if len(websites) > 0 {
		neverExpireDate, _ := time.Parse(constant.DateLayout, constant.DefaultDate)
		var wg sync.WaitGroup
		for _, site := range websites {
			if site.Status != constant.WebRunning || neverExpireDate.Equal(site.ExpireDate) {
				continue
			}
			expireDate, err := time.ParseInLocation(constant.DateLayout, site.ExpireDate.Format(constant.DateLayout), nyc)
			if err != nil {
				global.LOG.Errorf("time parse err %v", err)
				continue
			}
			if expireDate.Before(now) {
				wg.Add(1)
				go func(ws model.Website) {
					stopWebsite(ws.ID, ws.PrimaryDomain, &wg)
				}(site)
			}
		}
		wg.Wait()
	}
	global.LOG.Info("Website scheduled task has completed")
}

func stopWebsite(websiteId uint, websiteName string, wg *sync.WaitGroup) {
	websiteService := service.NewIWebsiteService()
	req := request.WebsiteOp{
		ID:      websiteId,
		Operate: constant.StopWeb,
	}
	if err := websiteService.OpWebsite(req); err != nil {
		global.LOG.Errorf("Website [%s]  seop failed err %v", websiteName, err)
	} else {
		global.LOG.Infof("Website [%s]  stopped successfully", websiteName)
	}
	wg.Done()
}
