package service

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
)

func (a AppService) UploadLocalAppPackage(packagePath, taskID string, overwrite bool) (string, []string, error) {
	appNames, err := inspectLocalAppPackage(packagePath)
	if err != nil {
		return "", nil, err
	}
	importTask, err := task.NewTaskWithOps(i18n.GetMsgByKey("LocalApp"), task.TaskImport, task.TaskScopeAppStore, taskID, 0)
	if err != nil {
		return "", nil, err
	}

	importTask.AddSubTask(
		task.GetTaskName(i18n.GetMsgByKey("LocalApp"), task.TaskImport, task.TaskScopeAppStore),
		func(t *task.Task) error {
			return a.importLocalAppPackageTask(t, packagePath, overwrite)
		},
		nil,
	)
	importTask.AddSubTask(
		task.GetTaskName(i18n.GetMsgByKey("LocalApp"), task.TaskSync, task.TaskScopeAppStore),
		func(t *task.Task) error {
			return a.syncLocalAppsTask(t)
		},
		nil,
	)

	go func() {
		if err := importTask.Execute(); err != nil {
			global.LOG.Errorf("import local app package failed: %v", err)
		}
	}()

	return importTask.Task.ID, appNames, nil
}

func (a AppService) importLocalAppPackageTask(taskItem *task.Task, packagePath string, overwrite bool) error {
	fileOp := files.NewFileOp()
	workDir := filepath.Dir(packagePath)
	extractDir := path.Join(workDir, "extract")
	defer func() {
		_ = fileOp.DeleteDir(workDir)
	}()

	if !strings.HasSuffix(strings.ToLower(packagePath), ".tar.gz") {
		return buserr.New("CustomAppStoreFileValid")
	}

	if fileOp.Stat(extractDir) {
		if err := fileOp.DeleteDir(extractDir); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(extractDir, constant.DirPerm); err != nil {
		return err
	}

	taskItem.LogStart("extract local app package")
	if err := fileOp.Decompress(packagePath, extractDir, files.SdkTarGz, ""); err != nil {
		taskItem.LogFailedWithErr("extract local app package", err)
		return err
	}
	taskItem.LogSuccess("extract local app package")

	appRoots, err := collectLocalAppRoots(extractDir)
	if err != nil {
		return err
	}
	if len(appRoots) == 0 {
		return fmt.Errorf("no valid 1Panel local app directories found in package")
	}

	for _, appRoot := range appRoots {
		if err := validateLocalAppRoot(appRoot); err != nil {
			return fmt.Errorf("invalid local app package %s: %w", filepath.Base(appRoot), err)
		}
	}
	for _, appRoot := range appRoots {
		appName := filepath.Base(appRoot)
		targetDir := path.Join(global.Dir.LocalAppResourceDir, appName)
		if fileOp.Stat(targetDir) && !overwrite {
			return fmt.Errorf("local app %s already exists", appName)
		}
	}

	for _, appRoot := range appRoots {
		appName := filepath.Base(appRoot)
		targetDir := path.Join(global.Dir.LocalAppResourceDir, appName)
		if fileOp.Stat(targetDir) {
			if !overwrite {
				return fmt.Errorf("local app %s already exists", appName)
			}
			if err := fileOp.DeleteDir(targetDir); err != nil {
				return err
			}
		}

		if err := moveLocalAppRoot(appRoot, targetDir); err != nil {
			return fmt.Errorf("move local app %s failed: %w", appName, err)
		}
		taskItem.LogSuccess(fmt.Sprintf("import local app package [%s]", appName))
	}

	return nil
}

func (a AppService) syncLocalAppsTask(t *task.Task) (err error) {
	fileOp := files.NewFileOp()
	localAppDir := global.Dir.LocalAppResourceDir
	if !fileOp.Stat(localAppDir) {
		return nil
	}

	dirEntries, err := os.ReadDir(localAppDir)
	if err != nil {
		return err
	}

	localApps := make([]model.App, 0)
	for _, dirEntry := range dirEntries {
		if !dirEntry.IsDir() {
			continue
		}
		appDir := filepath.Join(localAppDir, dirEntry.Name())
		appDirEntries, err := os.ReadDir(appDir)
		if err != nil {
			t.Log(i18n.GetWithNameAndErr("ErrAppDirNull", dirEntry.Name(), err))
			continue
		}
		app, err := handleLocalApp(appDir)
		if err != nil {
			t.Log(i18n.GetWithNameAndErr("LocalAppErr", dirEntry.Name(), err))
			continue
		}

		var appDetails []model.AppDetail
		for _, appDirEntry := range appDirEntries {
			if appDirEntry.IsDir() {
				appDetail := model.AppDetail{
					Version: appDirEntry.Name(),
					Status:  constant.AppNormal,
				}
				versionDir := filepath.Join(appDir, appDirEntry.Name())
				if err = handleLocalAppDetail(versionDir, &appDetail); err != nil {
					t.Log(i18n.GetMsgWithMap("LocalAppVersionErr", map[string]interface{}{"name": app.Name, "version": appDetail.Version, "err": err.Error()}))
					continue
				}
				appDetails = append(appDetails, appDetail)
			}
		}
		if len(appDetails) > 0 {
			app.Details = appDetails
			localApps = append(localApps, *app)
		} else {
			t.Log(i18n.GetWithName("LocalAppVersionNull", app.Name))
		}
	}

	var (
		newApps    []model.App
		deleteApps []model.App
		updateApps []model.App
		oldAppIds  []uint

		deleteAppIds     []uint
		deleteAppDetails []model.AppDetail
		newAppDetails    []model.AppDetail
		updateDetails    []model.AppDetail

		appTags []*model.AppTag
	)

	oldApps, _ := appRepo.GetBy(appRepo.WithResource(constant.AppResourceLocal))
	apps := make(map[string]model.App, len(oldApps))
	for _, old := range oldApps {
		old.Status = constant.AppTakeDown
		apps[old.Key] = old
	}
	for _, app := range localApps {
		if oldApp, ok := apps[app.Key]; ok {
			app.ID = oldApp.ID
			appDetails := make(map[string]model.AppDetail, len(oldApp.Details))
			for _, old := range oldApp.Details {
				old.Status = constant.AppTakeDown
				appDetails[old.Version] = old
			}
			for i, newDetail := range app.Details {
				version := newDetail.Version
				newDetail.Status = constant.AppNormal
				newDetail.AppId = app.ID
				oldDetail, exist := appDetails[version]
				if exist {
					newDetail.ID = oldDetail.ID
					delete(appDetails, version)
				}
				app.Details[i] = newDetail
			}
			for _, v := range appDetails {
				app.Details = append(app.Details, v)
			}
		}
		app.TagsKey = append(app.TagsKey, constant.AppResourceLocal)
		apps[app.Key] = app
	}

	for _, app := range apps {
		if app.ID == 0 {
			newApps = append(newApps, app)
		} else {
			oldAppIds = append(oldAppIds, app.ID)
			if app.Status == constant.AppTakeDown {
				installs, _ := appInstallRepo.ListBy(context.Background(), appInstallRepo.WithAppId(app.ID))
				if len(installs) > 0 {
					updateApps = append(updateApps, app)
					continue
				}
				deleteAppIds = append(deleteAppIds, app.ID)
				deleteApps = append(deleteApps, app)
				deleteAppDetails = append(deleteAppDetails, app.Details...)
			} else {
				updateApps = append(updateApps, app)
			}
		}
	}

	tags, _ := tagRepo.All()
	tagMap := make(map[string]uint, len(tags))
	for _, tag := range tags {
		tagMap[tag.Key] = tag.ID
	}

	tx, ctx := getTxAndContext()
	defer tx.Rollback()
	if len(newApps) > 0 {
		if err = appRepo.BatchCreate(ctx, newApps); err != nil {
			return err
		}
	}
	for _, update := range updateApps {
		if err = appRepo.Save(ctx, &update); err != nil {
			return err
		}
	}
	if len(deleteApps) > 0 {
		if err = appRepo.BatchDelete(ctx, deleteApps); err != nil {
			return err
		}
		if err = appDetailRepo.DeleteByAppIds(ctx, deleteAppIds); err != nil {
			return err
		}
	}

	if err = appTagRepo.DeleteByAppIds(ctx, oldAppIds); err != nil {
		return err
	}
	for _, newApp := range newApps {
		if newApp.ID > 0 {
			for _, detail := range newApp.Details {
				detail.AppId = newApp.ID
				newAppDetails = append(newAppDetails, detail)
			}
		}
	}
	for _, update := range updateApps {
		for _, detail := range update.Details {
			if detail.ID == 0 {
				detail.AppId = update.ID
				newAppDetails = append(newAppDetails, detail)
			} else {
				if detail.Status == constant.AppNormal {
					updateDetails = append(updateDetails, detail)
				} else {
					deleteAppDetails = append(deleteAppDetails, detail)
				}
			}
		}
	}

	allApps := append(newApps, updateApps...)
	for _, app := range allApps {
		for _, tagKey := range app.TagsKey {
			if tagId, ok := tagMap[tagKey]; ok {
				appTags = append(appTags, &model.AppTag{
					AppId: app.ID,
					TagId: tagId,
				})
			}
		}
	}

	if len(newAppDetails) > 0 {
		if err = appDetailRepo.BatchCreate(ctx, newAppDetails); err != nil {
			return err
		}
	}
	for _, updateAppDetail := range updateDetails {
		if err = appDetailRepo.Update(ctx, updateAppDetail); err != nil {
			return err
		}
	}
	if len(deleteAppDetails) > 0 {
		if err = appDetailRepo.BatchDelete(ctx, deleteAppDetails); err != nil {
			return err
		}
	}
	if len(oldAppIds) > 0 {
		if err = appTagRepo.DeleteByAppIds(ctx, oldAppIds); err != nil {
			return err
		}
	}
	if len(appTags) > 0 {
		if err = appTagRepo.BatchCreate(ctx, appTags); err != nil {
			return err
		}
	}
	tx.Commit()
	global.LOG.Infof("Synchronization of local applications completed")
	return nil
}

func collectLocalAppRoots(root string) ([]string, error) {
	var roots []string
	err := filepath.WalkDir(root, func(current string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !d.IsDir() {
			return nil
		}
		if current == root {
			return nil
		}
		if looksLikeLocalAppRoot(current) {
			roots = append(roots, current)
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(roots)
	return roots, nil
}

func looksLikeLocalAppRoot(dir string) bool {
	fileOp := files.NewFileOp()
	if !fileOp.Stat(path.Join(dir, "data.yml")) || !fileOp.Stat(path.Join(dir, "logo.png")) {
		return false
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if entry.IsDir() {
			versionDir := path.Join(dir, entry.Name())
			if fileOp.Stat(path.Join(versionDir, "data.yml")) && fileOp.Stat(path.Join(versionDir, "docker-compose.yml")) {
				return true
			}
		}
	}
	return false
}

func validateLocalAppRoot(appDir string) error {
	app, err := handleLocalApp(appDir)
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(appDir)
	if err != nil {
		return err
	}
	validVersions := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		appDetail := model.AppDetail{
			Version: entry.Name(),
			Status:  constant.AppNormal,
		}
		if err = handleLocalAppDetail(path.Join(appDir, entry.Name()), &appDetail); err == nil {
			validVersions++
		}
	}
	if validVersions == 0 {
		return fmt.Errorf("local app %s has no valid version directory", app.Name)
	}
	return nil
}

func moveLocalAppRoot(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	fileOp := files.NewFileOp()
	if err := fileOp.CopyDir(src, dst); err != nil {
		return err
	}
	return fileOp.DeleteDir(src)
}

func inspectLocalAppPackage(packagePath string) ([]string, error) {
	fileOp := files.NewFileOp()
	workDir := filepath.Dir(packagePath)
	inspectDir := path.Join(workDir, "inspect")
	defer func() {
		_ = fileOp.DeleteDir(inspectDir)
	}()

	if fileOp.Stat(inspectDir) {
		if err := fileOp.DeleteDir(inspectDir); err != nil {
			return nil, err
		}
	}
	if err := os.MkdirAll(inspectDir, constant.DirPerm); err != nil {
		return nil, err
	}
	if err := fileOp.Decompress(packagePath, inspectDir, files.SdkTarGz, ""); err != nil {
		return nil, err
	}

	appRoots, err := collectLocalAppRoots(inspectDir)
	if err != nil {
		return nil, err
	}
	if len(appRoots) == 0 {
		return nil, fmt.Errorf("no valid 1Panel local app directories found in package")
	}

	appNames := make([]string, 0, len(appRoots))
	for _, appRoot := range appRoots {
		if err := validateLocalAppRoot(appRoot); err != nil {
			return nil, fmt.Errorf("invalid local app package %s: %w", filepath.Base(appRoot), err)
		}
		appNames = append(appNames, filepath.Base(appRoot))
	}
	sort.Strings(appNames)
	return appNames, nil
}
