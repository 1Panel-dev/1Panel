package service

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"gorm.io/gorm"
)

func (w WebsiteService) BatchOpWebsite(req request.BatchWebsiteOp) error {
	if req.Operate == "clearProxyCache" {
		opTask, err := task.NewTaskWithOps(i18n.GetMsgByKey("ProxyCacheClear"), task.TaskBatch, task.TaskScopeWebsite, req.TaskID, 0)
		if err != nil {
			return err
		}
		opTask.AddSubTask("", func(t *task.Task) error {
			return batchClearProxyCache(t, req.IDs)
		}, nil)
		go func() { _ = opTask.Execute() }()
		return nil
	}
	websites, err := websiteRepo.List(repo.WithByIDs(req.IDs))
	if err != nil {
		return err
	}
	opTask, err := task.NewTaskWithOps(i18n.GetMsgByKey("Status"), task.TaskBatch, task.TaskScopeWebsite, req.TaskID, 0)
	if err != nil {
		return err
	}
	opWebsiteTask := func(t *task.Task) error { return w.batchOperateWebsites(t.TaskCtx, websites, req.Operate, t, nil) }
	opTask.AddSubTask("", opWebsiteTask, nil)

	go func() {
		_ = opTask.Execute()
	}()
	return nil
}

func batchClearProxyCache(t *task.Task, ids []uint) error {
	websites, err := websiteRepo.GetBy(repo.WithByIDs(ids))
	if err != nil {
		return err
	}
	// A nil entry is a requested website that no longer exists.
	pending := make(map[uint]*model.Website, len(ids))
	for _, id := range ids {
		pending[id] = nil
	}
	for i := range websites {
		pending[websites[i].ID] = &websites[i]
	}

	sitesDir := GetSiteDir("")
	var cleared, skipped, failed int
	var needsReload bool
	for _, id := range ids {
		web, ok := pending[id]
		if !ok {
			continue
		}
		delete(pending, id)
		if web == nil {
			t.LogFailedWithErr(fmt.Sprintf("ID %d", id), gorm.ErrRecordNotFound)
			failed++
			continue
		}
		changed, err := cleanWebsiteProxyCache(sitesDir, web.Alias)
		needsReload = needsReload || changed
		switch {
		case err != nil:
			t.LogFailedWithErr(web.PrimaryDomain, err)
			failed++
		case changed:
			t.LogSuccess(web.PrimaryDomain)
			cleared++
		default:
			t.Log(web.PrimaryDomain + " · " + i18n.GetMsgByKey("ProxyCacheEmpty"))
			skipped++
		}
	}
	summary := i18n.GetMsgWithMap("ProxyCacheSummary", map[string]interface{}{"cleared": cleared, "skipped": skipped, "failed": failed})
	t.Log(summary)
	if needsReload {
		if err := reloadProxyCacheNginx(); err != nil {
			return fmt.Errorf("%s: %w", i18n.GetMsgByKey("ProxyCacheReloadFailed"), err)
		}
	}
	if failed > 0 {
		return errors.New(summary)
	}
	return nil
}

func (w WebsiteService) BatchSetGroup(req request.BatchWebsiteGroup) error {
	websites, err := websiteRepo.List(repo.WithByIDs(req.IDs))
	if err != nil {
		return err
	}
	for _, web := range websites {
		web.WebsiteGroupID = req.GroupID
		// Persist the completed reload even if the execution deadline has just passed.
		if err := websiteRepo.Save(context.Background(), &web); err != nil {
			return err
		}
	}
	return nil
}

func (w WebsiteService) BatchSetHttps(ctx context.Context, req request.BatchWebsiteHttps) error {
	websites, err := websiteRepo.List(repo.WithByIDs(req.IDs))
	if err != nil {
		return err
	}
	opTask, err := task.NewTaskWithOps(i18n.GetMsgByKey("SSL"), task.TaskBatch, task.TaskScopeWebsite, req.TaskID, 0)
	if err != nil {
		return err
	}
	websiteHttpsOp := request.WebsiteHTTPSOp{
		Enable:                true,
		WebsiteSSLID:          req.WebsiteSSLID,
		Type:                  req.Type,
		PrivateKey:            req.PrivateKey,
		Certificate:           req.Certificate,
		PrivateKeyPath:        req.PrivateKeyPath,
		CertificatePath:       req.CertificatePath,
		ImportType:            req.ImportType,
		HttpConfig:            req.HttpConfig,
		SSLProtocol:           req.SSLProtocol,
		Algorithm:             req.Algorithm,
		Hsts:                  req.Hsts,
		HstsIncludeSubDomains: req.HstsIncludeSubDomains,
		HttpsPorts:            req.HttpsPorts,
		Http3:                 req.Http3,
	}
	if req.Type == constant.SSLManual {
		websiteSSL, err := getManualWebsiteSSL(websiteHttpsOp)
		if err != nil {
			return err
		}
		if err = websiteSSLRepo.Create(ctx, &websiteSSL); err != nil {
			return err
		}
		websiteHttpsOp.Type = constant.SSLExisted
		websiteHttpsOp.WebsiteSSLID = websiteSSL.ID
	}
	opWebsiteTask := func(t *task.Task) error {
		for _, web := range websites {
			if web.Type == constant.Stream {
				continue
			}
			websiteHttpsOp.WebsiteID = web.ID
			msg := fmt.Sprintf("%s [%s] %s", i18n.GetMsgByKey("Set"), web.PrimaryDomain, i18n.GetMsgByKey("SSL"))
			if _, err := w.OpWebsiteHTTPS(ctx, websiteHttpsOp); err != nil {
				t.LogFailedWithErr(msg, err)
				continue
			}
			t.LogSuccess(msg)
		}
		return nil
	}

	opTask.AddSubTask("", opWebsiteTask, nil)
	go func() {
		_ = opTask.Execute()
	}()
	return nil
}

func (w WebsiteService) batchOperateWebsites(ctx context.Context, websites []model.Website, operate string, t *task.Task, completed map[uint]bool) error {
	sort.SliceStable(websites, func(i, j int) bool {
		if websites[i].Type == constant.Subsite && websites[j].Type != constant.Subsite {
			return true
		}
		if websites[i].Type != constant.Subsite && websites[j].Type == constant.Subsite {
			return false
		}
		return false
	})

	var failures []error
	for _, web := range websites {
		if completed[web.ID] {
			continue
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		msg := fmt.Sprintf("%s %s", i18n.GetMsgByKey(operate), web.PrimaryDomain)
		switch operate {
		case constant.StopWeb, constant.StartWeb:
			if err := opWebsiteContext(ctx, &web, operate); err != nil {
				t.LogFailedWithErr(msg, err)
				failures = append(failures, err)
				continue
			}
			// Persist the completed reload even if the execution deadline has just passed.
			if err := websiteRepo.Save(context.Background(), &web); err != nil {
				failures = append(failures, err)
				continue
			}
		case "delete":
			if err := w.DeleteWebsite(request.WebsiteDelete{
				ID: web.ID,
			}); err != nil {
				t.LogFailedWithErr(msg, err)
				failures = append(failures, err)
				continue
			}
		}

		if completed != nil {
			completed[web.ID] = true
		}
		t.LogSuccess(msg)
	}

	return errors.Join(failures...)
}
