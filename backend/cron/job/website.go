package job

import (
	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/app/repo"
	"github.com/1Panel-dev/1Panel/backend/app/service"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"sync"
	"time"
)

type website struct{}

func NewWebsiteJob() *website {
	return &website{}
}

func (w *website) Run() {
	websites, _ := repo.NewIWebsiteRepo().List()
	global.LOG.Info("website cron job start...")
	now := time.Now().Add(10 * time.Second)
	if len(websites) > 0 {
		neverExpireDate, _ := time.Parse(constant.DateLayout, constant.DefaultDate)
		var wg sync.WaitGroup
		for _, site := range websites {
			if site.Status != constant.WebRunning || neverExpireDate.Equal(site.ExpireDate) {
				continue
			}
			expireDate, _ := time.ParseInLocation(constant.DateTimeLayout, site.ExpireDate.String(), time.Now().Location())
			if expireDate.Before(now) {
				wg.Add(1)
				go func(ws model.Website) {
					stopWebsite(ws.ID, &wg)
				}(site)
			}
		}
		wg.Wait()
	}
	global.LOG.Info("website cron job end...")
}

func stopWebsite(websiteId uint, wg *sync.WaitGroup) {
	websiteService := service.NewWebsiteService()
	req := request.WebsiteOp{
		ID:      websiteId,
		Operate: constant.StopWeb,
	}
	if err := websiteService.OpWebsite(req); err != nil {
		global.LOG.Errorf("stop website err: %s", err.Error())
	}
	wg.Done()
}
