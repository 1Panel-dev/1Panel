package service

import (
	"encoding/base64"
	"encoding/json"
	"github.com/1Panel-dev/1Panel/app/dto"
	"github.com/1Panel-dev/1Panel/app/model"
	"github.com/1Panel-dev/1Panel/app/repo"
	"github.com/1Panel-dev/1Panel/global"
	"golang.org/x/net/context"
	"os"
	"path"
	"reflect"
)

type AppService struct {
}

func (a AppService) Page(req dto.AppRequest) (interface{}, error) {

	var opts []repo.DBOption
	opts = append(opts, commonRepo.WithOrderBy("name"))
	if req.Name != "" {
		opts = append(opts, commonRepo.WithLikeName(req.Name))
	}
	if len(req.Types) != 0 {
		opts = append(opts, appRepo.WithInTypes(req.Types))
	}
	var res dto.AppRes
	total, apps, err := appRepo.Page(req.Page, req.PageSize, opts...)
	if err != nil {
		return nil, err
	}
	var appDTOs []*dto.AppDTO
	for _, a := range apps {
		appDTO := &dto.AppDTO{
			App: a,
		}
		appDTOs = append(appDTOs, appDTO)
		appTags, err := appTagRepo.GetByAppId(a.ID)
		if err != nil {
			continue
		}
		var tagIds []uint
		for _, at := range appTags {
			tagIds = append(tagIds, at.TagId)
		}
		tags, err := tagRepo.GetByIds(tagIds)
		if err != nil {
			continue
		}
		appDTO.Tags = tags
	}
	res.Items = appDTOs
	res.Total = total
	tags, err := tagRepo.All()
	if err != nil {
		return nil, err
	}
	res.Tags = tags

	return res, nil
}

func (a AppService) Sync() error {
	//TODO 从 oss 拉取最新列表
	var appConfig model.AppConfig
	appConfig.OssPath = global.CONF.System.AppOss

	appDir := path.Join(global.CONF.System.ResourceDir, "apps")
	iconDir := path.Join(appDir, "icons")
	listFile := path.Join(appDir, "list.json")

	content, err := os.ReadFile(listFile)
	if err != nil {
		return err
	}
	list := &dto.AppList{}
	if err := json.Unmarshal(content, list); err != nil {
		return err
	}
	appConfig.Version = list.Version
	appConfig.CanUpdate = false

	var (
		tags       []*model.Tag
		addApps    []*model.App
		updateApps []*model.App
		appTags    []*model.AppTag
	)

	for _, t := range list.Tags {
		tags = append(tags, &model.Tag{
			Key:  t.Key,
			Name: t.Name,
		})
	}

	db := global.DB
	dbCtx := context.WithValue(context.Background(), "db", db)
	for _, l := range list.Items {
		icon, err := os.ReadFile(path.Join(iconDir, l.Icon))
		if err != nil {
			global.LOG.Errorf("get [%s] icon error: %s", l.Name, err.Error())
			continue
		}
		iconStr := base64.StdEncoding.EncodeToString(icon)
		app := &model.App{
			Name:      l.Name,
			Key:       l.Key,
			ShortDesc: l.ShortDesc,
			Author:    l.Author,
			Source:    l.Source,
			Icon:      iconStr,
			Type:      l.Type,
		}
		app.TagsKey = l.Tags
		old, _ := appRepo.GetByKey(dbCtx, l.Key)
		if reflect.DeepEqual(old, &model.App{}) {
			addApps = append(addApps, app)
		} else {
			app.ID = old.ID
			updateApps = append(updateApps, app)
		}

		versions := l.Versions
		for _, v := range versions {
			detail := &model.AppDetail{
				Version: v,
			}
			detailPath := path.Join(appDir, l.Key, v)
			if _, err := os.Stat(detailPath); err != nil {
				global.LOG.Errorf("get [%s] folder error: %s", detailPath, err.Error())
				continue
			}
			readmeStr, err := os.ReadFile(path.Join(detailPath, "README.md"))
			if err != nil {
				global.LOG.Errorf("get [%s] README error: %s", detailPath, err.Error())
			}
			detail.Readme = string(readmeStr)
			dockerComposeStr, err := os.ReadFile(path.Join(detailPath, "docker-compose.yml"))
			if err != nil {
				global.LOG.Errorf("get [%s] docker-compose.yml error: %s", detailPath, err.Error())
				continue
			}
			detail.DockerCompose = string(dockerComposeStr)
			formStr, err := os.ReadFile(path.Join(detailPath, "form.json"))
			if err != nil {
				global.LOG.Errorf("get [%s] form.json error: %s", detailPath, err.Error())
			}
			detail.FormFields = string(formStr)
			app.Details = append(app.Details, detail)
		}
	}
	tx := global.DB.Begin()
	ctx := context.WithValue(context.Background(), "db", tx)
	if len(addApps) > 0 {
		if err := appRepo.BatchCreate(ctx, addApps); err != nil {
			tx.Rollback()
			return err
		}
	}
	if err := tagRepo.DeleteAll(ctx); err != nil {
		tx.Rollback()
		return err
	}

	if len(tags) > 0 {
		if err := tagRepo.BatchCreate(ctx, tags); err != nil {
			tx.Rollback()
			return err
		}
	}

	tagMap := make(map[string]uint, len(tags))

	for _, t := range tags {
		tagMap[t.Key] = t.ID
	}

	for _, a := range updateApps {
		if err := appRepo.Save(ctx, a); err != nil {
			tx.Rollback()
			return err
		}
	}
	apps := append(addApps, updateApps...)

	var (
		appDetails []*model.AppDetail
		appIds     []uint
	)
	for _, a := range apps {

		for _, t := range a.TagsKey {
			tagId, ok := tagMap[t]
			if ok {
				appTags = append(appTags, &model.AppTag{
					AppId: a.ID,
					TagId: tagId,
				})
			}
		}

		for _, d := range a.Details {
			d.AppId = a.ID
			appDetails = append(appDetails, d)
		}
		appIds = append(appIds, a.ID)
	}

	if err := appDetailRepo.DeleteByAppIds(ctx, appIds); err != nil {
		tx.Rollback()
		return err
	}

	if len(appDetails) > 0 {
		if err := appDetailRepo.BatchCreate(ctx, appDetails); err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := appTagRepo.DeleteByAppIds(ctx, appIds); err != nil {
		tx.Rollback()
		return err
	}

	if len(appTags) > 0 {
		if err := appTagRepo.BatchCreate(ctx, appTags); err != nil {
			tx.Rollback()
			return err
		}
	}

	tx.Commit()
	return nil
}
