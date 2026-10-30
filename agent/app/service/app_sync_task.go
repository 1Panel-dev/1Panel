package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
)

type appSyncContext struct {
	task           *task.Task
	baseRemoteUrl  string
	systemVersion  string
	appsMap        map[string]model.App
	settingService ISettingService
	list           *dto.AppList
	oldAppIds      []uint
	appTags        []*model.AppTag
	skipMetaSync   bool
}

func (a AppService) createSyncAppStoreTask(sharedCtx **appSyncContext) func(t *task.Task) error {
	return func(t *task.Task) (err error) {
		t.LogStart(i18n.GetMsgByKey("AppStore") + " " + i18n.GetMsgByKey("TaskSync"))

		updateRes, err := a.GetAppUpdate()
		if err != nil {
			t.LogFailedWithErr(i18n.GetMsgByKey("CheckAppStoreUpdate"), err)
			return err
		}
		if !updateRes.CanUpdate {
			if updateRes.IsSyncing {
				t.Log(i18n.GetMsgByKey("AppStoreIsSyncing"))
				return nil
			}
			if err := refreshOfficialAppIcons(updateRes.AppList, t); err != nil {
				t.LogFailedWithErr(i18n.GetMsgByKey("PersistApps"), err)
				return err
			}
			global.LOG.Infof("[AppStore] Appstore is up to date")
			t.Log(i18n.GetMsgByKey("AppStoreIsUpToDate"))
			*sharedCtx = &appSyncContext{skipMetaSync: true}
			t.LogSuccess(i18n.GetMsgByKey("AppStore") + " " + i18n.GetMsgByKey("TaskSync"))
			return nil
		}

		list := &dto.AppList{}
		if updateRes.AppList == nil {
			list, err = getAppList()
			if err != nil {
				t.LogFailedWithErr(i18n.GetMsgByKey("DownloadAppList"), err)
				return err
			}
		} else {
			list = updateRes.AppList
		}

		settingService := NewISettingService()
		if err := settingService.Update("AppStoreSyncStatus", constant.StatusSyncing); err != nil {
			global.LOG.Warnf("[AppStore] failed to update sync status to syncing: %v", err)
		}

		setting, err := settingService.GetSettingInfo()
		if err != nil {
			t.LogFailedWithErr("GetSettingInfo", err)
			return err
		}

		ctx := &appSyncContext{
			task:           t,
			baseRemoteUrl:  fmt.Sprintf("%s/%s/1panel", global.AppRepoURL(), global.CONF.Base.Mode),
			systemVersion:  setting.SystemVersion,
			settingService: settingService,
			list:           list,
			appTags:        make([]*model.AppTag, 0),
		}

		if err = SyncTags(list.Extra); err != nil {
			t.LogFailedWithErr(i18n.GetMsgByKey("SyncTags"), err)
			return err
		}
		deleteCustomApp()

		oldApps, err := appRepo.GetBy(appRepo.WithNotLocal())
		if err != nil {
			t.LogFailedWithErr(i18n.GetMsgByKey("LoadLocalApps"), err)
			return err
		}
		ctx.oldAppIds = make([]uint, 0, len(oldApps))
		for _, old := range oldApps {
			ctx.oldAppIds = append(ctx.oldAppIds, old.ID)
		}

		ctx.appsMap = getApps(oldApps, list.Apps, setting.SystemVersion, t)

		var addCount, updateCount, deleteCount int
		if err = ctx.classifyAndPersistAppsWithStats(&addCount, &updateCount, &deleteCount); err != nil {
			t.LogFailedWithErr(i18n.GetMsgByKey("PersistApps"), err)
			return err
		}

		if err := settingService.Update("AppStoreSyncStatus", constant.StatusSyncSuccess); err != nil {
			global.LOG.Warnf("[AppStore] failed to update sync status to success: %v", err)
		}
		if err := settingService.Update("AppStoreLastModified", strconv.Itoa(list.LastModified)); err != nil {
			global.LOG.Warnf("[AppStore] failed to update last modified: %v", err)
		}
		global.LOG.Infof("[AppStore] Appstore sync completed")

		*sharedCtx = ctx
		t.LogSuccess(i18n.GetMsgByKey("AppStore") + " " + i18n.GetMsgByKey("TaskSync"))
		return nil
	}
}

func refreshOfficialAppIcons(list *dto.AppList, t *task.Task) error {
	if list == nil {
		return nil
	}
	apps, err := appRepo.GetBy(appRepo.WithResource(constant.AppResourceRemote))
	if err != nil {
		return err
	}
	systemVersion, err := settingRepo.GetValueByKey("SystemVersion")
	if err != nil {
		return err
	}
	updated := getApps(apps, list.Apps, systemVersion, t)
	for _, app := range apps {
		candidate := updated[app.Key]
		if app.Status != constant.AppNormal || candidate.Status != constant.AppNormal || candidate.Icon == app.Icon {
			continue
		}
		if err := appRepo.UpdateIcon(app.ID, candidate.Icon); err != nil {
			return err
		}
	}
	return nil
}

type appWorkItem struct {
	appDef dto.AppDefine
	app    model.App
}

func (c *appSyncContext) processOneApp(item appWorkItem) model.App {
	app := item.app
	l := item.appDef

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

		downloadURL := fmt.Sprintf("%s/%s", versionUrl, app.Key+"-"+version+".tar.gz")
		if v.LastModified <= 0 || detail.LastModified != v.LastModified || detail.Params != string(paramByte) || detail.DownloadUrl != downloadURL {
			detail.DockerCompose = ""
		}
		detail.Params = string(paramByte)
		detail.DownloadUrl = downloadURL
		detail.DownloadCallBackUrl = v.DownloadCallBackUrl
		detail.Update = true
		detail.LastModified = v.LastModified
		detailsMap[version] = detail
	}

	newDetails := make([]model.AppDetail, 0, len(detailsMap))
	for _, detail := range detailsMap {
		newDetails = append(newDetails, detail)
	}
	app.Details = newDetails

	return app
}

func (c *appSyncContext) syncAppDetails() error {
	total := len(c.list.Apps)
	global.LOG.Infof("[AppStore] sync app detail start, total apps: %d", total)
	for i, definition := range c.list.Apps {
		app, ok := c.appsMap[definition.AppProperty.Key]
		if ok {
			c.appsMap[app.Key] = c.processOneApp(appWorkItem{appDef: definition, app: app})
		}
		if (i+1)%max(1, total/4) == 0 || i+1 == total {
			c.task.LogWithProgress(i18n.GetMsgByKey("SyncAppDetail"), i+1, total)
		}
	}
	global.LOG.Infof("[AppStore] app detail sync completed, total: %d", total)
	return nil
}

func (a AppService) createSyncAppStoreMetaTask(sharedCtx **appSyncContext) func(t *task.Task) error {
	return func(t *task.Task) (err error) {
		t.LogStart(i18n.GetMsgByKey("SyncAppDetail"))
		ctx := *sharedCtx
		if ctx == nil {
			global.LOG.Warnf("[AppStore] meta sync skipped: shared context is nil")
			t.Log(i18n.GetMsgByKey("SyncAppDetail") + " skipped: shared context is nil")
			return nil
		}

		if ctx.skipMetaSync {
			global.LOG.Infof("[AppStore] meta sync skipped: no update needed")
			t.Log(i18n.GetMsgByKey("SyncAppDetail") + " skipped: no update needed")
			return nil
		}

		if ctx.list == nil || ctx.appsMap == nil {
			global.LOG.Errorf("[AppStore] meta sync failed: shared context data not initialized")
			err := fmt.Errorf("shared context data not initialized")
			t.LogFailedWithErr(i18n.GetMsgByKey("SyncAppDetail"), err)
			return err
		}

		t.Logf("%s: %d apps", i18n.GetMsgByKey("SyncAppDetail"), len(ctx.list.Apps))

		ctx.task = t
		ctx.appTags = make([]*model.AppTag, 0)

		if err = ctx.syncAppDetails(); err != nil {
			t.LogFailedWithErr(i18n.GetMsgByKey("SyncAppDetail"), err)
			return err
		}

		if err = ctx.classifyAndPersistApps(); err != nil {
			t.LogFailedWithErr(i18n.GetMsgByKey("PersistAppDetails"), err)
			return err
		}

		global.LOG.Infof("[AppStore] Appstore meta sync completed")
		return nil
	}
}

func (c *appSyncContext) classifyAndPersistApps() (err error) {
	var addCount, updateCount, deleteCount int
	return c.classifyAndPersistAppsWithStats(&addCount, &updateCount, &deleteCount)
}

func (c *appSyncContext) classifyAndPersistAppsWithStats(addCount, updateCount, deleteCount *int) (err error) {
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

	if len(addAppArray) > 0 {
		addKeys := make([]string, 0, len(addAppArray))
		for _, app := range addAppArray {
			addKeys = append(addKeys, app.Key)
		}
		existingApps, _ := appRepo.GetBy(appRepo.WithKeyIn(addKeys))
		if len(existingApps) > 0 {
			existingMap := make(map[string]model.App, len(existingApps))
			for _, e := range existingApps {
				existingMap[e.Key] = e
			}
			filteredAdd := make([]model.App, 0, len(addAppArray))
			for _, app := range addAppArray {
				if existing, ok := existingMap[app.Key]; ok {
					app.ID = existing.ID
					if len(app.Details) == 0 {
						app.Details = existing.Details
					}
					updateAppArray = append(updateAppArray, app)
				} else {
					filteredAdd = append(filteredAdd, app)
				}
			}
			addAppArray = filteredAdd
		}
	}

	*addCount = len(addAppArray)
	*updateCount = len(updateAppArray)
	*deleteCount = len(deleteAppArray)

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

	if len(updateAppArray) > 0 {
		for _, update := range updateAppArray {
			if err = appRepo.Save(ctx, &update); err != nil {
				return
			}
		}
	}

	apps := append(addAppArray, updateAppArray...)

	var (
		addDetails    []model.AppDetail
		updateDetails []model.AppDetail
		deleteDetails []model.AppDetail
	)

	totalDetails := 0
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
			totalDetails++
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

	if len(updateDetails) > 0 {
		for _, u := range updateDetails {
			if err = appDetailRepo.Update(ctx, u); err != nil {
				return
			}
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

	if err = tx.Commit().Error; err != nil {
		return
	}
	return nil
}
