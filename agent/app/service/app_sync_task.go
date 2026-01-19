package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/req_helper"
	"github.com/1Panel-dev/1Panel/agent/utils/xpack"
)

type appSyncContext struct {
	task           *task.Task
	httpClient     http.Client
	baseRemoteUrl  string
	systemVersion  string
	appsMap        map[string]model.App
	settingService ISettingService
	list           *dto.AppList
	oldAppIds      []uint
	appTags        []*model.AppTag
}

func (a AppService) syncAppStoreTask(t *task.Task) (err error) {
	updateRes, err := a.GetAppUpdate()
	if err != nil {
		return err
	}
	if !updateRes.CanUpdate {
		if updateRes.IsSyncing {
			t.Log(i18n.GetMsgByKey("AppStoreIsSyncing"))
			return nil
		}
		global.LOG.Infof("[AppStore] Appstore is up to date")
		t.Log(i18n.GetMsgByKey("AppStoreIsUpToDate"))
		return nil
	}

	list := &dto.AppList{}
	if updateRes.AppList == nil {
		list, err = getAppList()
		if err != nil {
			return err
		}
	} else {
		list = updateRes.AppList
	}

	settingService := NewISettingService()
	_ = settingService.Update("AppStoreSyncStatus", constant.StatusSyncing)

	setting, err := settingService.GetSettingInfo()
	if err != nil {
		return err
	}

	ctx := &appSyncContext{
		task:           t,
		httpClient:     http.Client{Timeout: time.Duration(constant.TimeOut20s) * time.Second, Transport: xpack.LoadRequestTransport()},
		baseRemoteUrl:  fmt.Sprintf("%s/%s/1panel", global.CONF.RemoteURL.AppRepo, global.CONF.Base.Mode),
		systemVersion:  setting.SystemVersion,
		settingService: settingService,
		list:           list,
		appTags:        make([]*model.AppTag, 0),
	}

	if err = SyncTags(list.Extra); err != nil {
		return err
	}
	deleteCustomApp()

	oldApps, err := appRepo.GetBy(appRepo.WithNotLocal())
	if err != nil {
		return err
	}
	ctx.oldAppIds = make([]uint, 0, len(oldApps))
	for _, old := range oldApps {
		ctx.oldAppIds = append(ctx.oldAppIds, old.ID)
	}

	ctx.appsMap = getApps(oldApps, list.Apps, setting.SystemVersion, t)

	if err = ctx.syncAppIconsAndDetails(); err != nil {
		return err
	}

	if err = ctx.classifyAndPersistApps(); err != nil {
		return err
	}

	_ = settingService.Update("AppStoreSyncStatus", constant.StatusSyncSuccess)
	_ = settingService.Update("AppStoreLastModified", strconv.Itoa(list.LastModified))
	global.LOG.Infof("[AppStore] Appstore sync completed")
	return nil
}

func (c *appSyncContext) syncAppIconsAndDetails() error {
	c.task.LogStart(i18n.GetMsgByKey("SyncAppDetail"))
	global.LOG.Infof("[AppStore] sync app detail start, total: %d", len(c.list.Apps))

	downloadIconNum := 0
	total := len(c.list.Apps)

	for _, l := range c.list.Apps {
		downloadIconNum++
		if downloadIconNum%10 == 0 {
			c.task.LogWithProgress(i18n.GetMsgByKey("SyncAppDetail"), downloadIconNum, total)
		}

		app, ok := c.appsMap[l.AppProperty.Key]
		if !ok {
			continue
		}

		iconStr := c.downloadAppIcon(l.Icon)
		if iconStr == "" {
			global.LOG.Infof("[AppStore] save failed url=%s", l.Icon)
		}
		app.Icon = iconStr

		app.TagsKey = l.AppProperty.Tags
		if l.AppProperty.Recommend > 0 {
			app.Recommend = l.AppProperty.Recommend
		} else {
			app.Recommend = 9999
		}
		app.ReadMe = l.ReadMe
		app.LastModified = l.LastModified

		versions := l.Versions
		detailsMap := getAppDetails(app.Details, versions)
		for _, v := range versions {
			version := v.Name
			detail := detailsMap[version]
			versionUrl := fmt.Sprintf("%s/%s/%s", c.baseRemoteUrl, app.Key, version)

			paramByte, _ := json.Marshal(v.AppForm)
			var appForm dto.AppForm
			_ = json.Unmarshal(paramByte, &appForm)

			if appForm.SupportVersion > 0 && common.CompareVersion(strconv.FormatFloat(appForm.SupportVersion, 'f', -1, 64), c.systemVersion) {
				delete(detailsMap, version)
				continue
			}

			if _, ok := InitTypes[app.Type]; ok {
				dockerComposeUrl := fmt.Sprintf("%s/%s", versionUrl, "docker-compose.yml")
				_, composeRes, err := req_helper.HandleRequestWithClient(&c.httpClient, dockerComposeUrl, http.MethodGet, constant.TimeOut20s)
				if err == nil {
					detail.DockerCompose = string(composeRes)
				}
			} else {
				detail.DockerCompose = ""
			}

			detail.Params = string(paramByte)
			detail.DownloadUrl = fmt.Sprintf("%s/%s", versionUrl, app.Key+"-"+version+".tar.gz")
			detail.DownloadCallBackUrl = v.DownloadCallBackUrl
			detail.Update = true
			detail.LastModified = v.LastModified
			detailsMap[version] = detail
		}

		var newDetails []model.AppDetail
		for _, detail := range detailsMap {
			newDetails = append(newDetails, detail)
		}
		app.Details = newDetails
		c.appsMap[l.AppProperty.Key] = app
	}

	global.LOG.Infof("[AppStore] download icon success: %d, total: %d",
		downloadIconNum, total)

	c.task.LogSuccess(i18n.GetMsgByKey("SyncAppDetail"))
	return nil
}

func (c *appSyncContext) downloadAppIcon(iconUrl string) string {
	iconStr := ""

	code, iconRes, err := req_helper.HandleRequestWithClient(&c.httpClient, iconUrl, http.MethodGet, constant.TimeOut20s)
	if err == nil {
		if code == http.StatusOK {
			if len(iconRes) > 0 {
				if iconRes[0] != '<' {
					iconStr = base64.StdEncoding.EncodeToString(iconRes)
				}
			}
		} else {
			global.LOG.Infof("[AppStore] download failed status=%d", code)
		}
	}
	return iconStr
}

func (c *appSyncContext) classifyAndPersistApps() (err error) {
	tags, _ := tagRepo.All()
	var (
		addAppArray    []model.App
		updateAppArray []model.App
		deleteAppArray []model.App
		deleteIds      []uint
		tagMap         = make(map[string]uint, len(tags))
	)

	for _, v := range c.appsMap {
		if v.ID == 0 {
			addAppArray = append(addAppArray, v)
		} else {
			if v.Status == constant.AppTakeDown {
				installs, _ := appInstallRepo.ListBy(context.Background(), appInstallRepo.WithAppId(v.ID))
				if len(installs) > 0 {
					updateAppArray = append(updateAppArray, v)
					continue
				}
				deleteAppArray = append(deleteAppArray, v)
				deleteIds = append(deleteIds, v.ID)
			} else {
				updateAppArray = append(updateAppArray, v)
			}
		}
	}

	tx, ctx := getTxAndContext()
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
	}()

	if len(addAppArray) > 0 {
		if err = appRepo.BatchCreate(ctx, addAppArray); err != nil {
			return
		}
	}

	if len(deleteAppArray) > 0 {
		if err = appRepo.BatchDelete(ctx, deleteAppArray); err != nil {
			return
		}
		if err = appDetailRepo.DeleteByAppIds(ctx, deleteIds); err != nil {
			return
		}
	}

	for _, tag := range tags {
		tagMap[tag.Key] = tag.ID
	}

	for _, update := range updateAppArray {
		if err = appRepo.Save(ctx, &update); err != nil {
			return
		}
	}

	apps := append(addAppArray, updateAppArray...)

	var (
		addDetails    []model.AppDetail
		updateDetails []model.AppDetail
		deleteDetails []model.AppDetail
	)

	for _, app := range apps {
		for _, tag := range app.TagsKey {
			tagId, ok := tagMap[tag]
			if ok {
				exist, _ := appTagRepo.GetFirst(ctx, appTagRepo.WithByTagID(tagId), appTagRepo.WithByAppID(app.ID))
				if exist == nil {
					c.appTags = append(c.appTags, &model.AppTag{
						AppId: app.ID,
						TagId: tagId,
					})
				}
			}
		}

		for _, d := range app.Details {
			d.AppId = app.ID
			if d.ID == 0 {
				addDetails = append(addDetails, d)
			} else {
				if d.Status == constant.AppTakeDown {
					runtime, _ := runtimeRepo.GetFirst(ctx, runtimeRepo.WithDetailId(d.ID))
					if runtime != nil {
						updateDetails = append(updateDetails, d)
						continue
					}
					installs, _ := appInstallRepo.ListBy(ctx, appInstallRepo.WithDetailIdsIn([]uint{d.ID}))
					if len(installs) > 0 {
						updateDetails = append(updateDetails, d)
						continue
					}
					deleteDetails = append(deleteDetails, d)
				} else {
					updateDetails = append(updateDetails, d)
				}
			}
		}
	}

	if len(addDetails) > 0 {
		if err = appDetailRepo.BatchCreate(ctx, addDetails); err != nil {
			return
		}
	}

	if len(deleteDetails) > 0 {
		if err = appDetailRepo.BatchDelete(ctx, deleteDetails); err != nil {
			return
		}
	}

	for _, u := range updateDetails {
		if err = appDetailRepo.Update(ctx, u); err != nil {
			return
		}
	}

	if len(c.oldAppIds) > 0 {
		if err = appTagRepo.DeleteByAppIds(ctx, deleteIds); err != nil {
			return
		}
	}

	if len(c.appTags) > 0 {
		if err = appTagRepo.BatchCreate(ctx, c.appTags); err != nil {
			return
		}
	}

	tx.Commit()
	return nil
}
