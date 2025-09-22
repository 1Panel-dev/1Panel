package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/core/app/dto"
	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/app/task"
	"github.com/1Panel-dev/1Panel/core/buserr"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/i18n"
	"github.com/1Panel-dev/1Panel/core/utils/common"
	"github.com/1Panel-dev/1Panel/core/utils/files"
	"github.com/1Panel-dev/1Panel/core/utils/req_helper"
	"github.com/1Panel-dev/1Panel/core/utils/xpack"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"gopkg.in/yaml.v2"
)

type ScriptService struct{}

type IScriptService interface {
	Search(ctx *gin.Context, req dto.SearchPageWithGroup) (int64, interface{}, error)
	Create(req dto.ScriptOperate) error
	Update(req dto.ScriptOperate) error
	Delete(ids dto.OperateByIDs) error
	Sync(req dto.OperateByTaskID) error
}

func NewIScriptService() IScriptService {
	return &ScriptService{}
}

func (u *ScriptService) Search(ctx *gin.Context, req dto.SearchPageWithGroup) (int64, interface{}, error) {
	options := []global.DBOption{repo.WithOrderBy("created_at desc")}
	if len(req.Info) != 0 {
		options = append(options, scriptRepo.WithByInfo(req.Info))
	}
	list, err := scriptRepo.GetList(options...)
	if err != nil {
		return 0, nil, err
	}
	groups, _ := groupRepo.GetList(repo.WithByType("script"))
	groupMap := make(map[uint]string)
	for _, item := range groups {
		groupMap[item.ID] = item.Name
	}
	var data []dto.ScriptInfo
	for _, itemData := range list {
		var item dto.ScriptInfo
		if err := copier.Copy(&item, &itemData); err != nil {
			global.LOG.Errorf("copy scripts to dto backup info failed, err: %v", err)
		}
		if item.IsSystem {
			lang := strings.ToLower(common.GetLang(ctx))
			var nameMap = make(map[string]string)
			_ = json.Unmarshal([]byte(item.Name), &nameMap)
			var descriptionMap = make(map[string]string)
			_ = json.Unmarshal([]byte(item.Description), &descriptionMap)
			if val, ok := nameMap[lang]; ok {
				item.Name = val
			}
			if val, ok := descriptionMap[lang]; ok {
				item.Description = val
			}
		}
		matchGroup := false
		groupIDs := strings.Split(itemData.Groups, ",")
		for _, idItem := range groupIDs {
			id, _ := strconv.Atoi(idItem)
			if id == 0 {
				continue
			}
			if uint(id) == req.GroupID {
				matchGroup = true
			}
			item.GroupList = append(item.GroupList, uint(id))
			item.GroupBelong = append(item.GroupBelong, groupMap[uint(id)])
		}
		if req.GroupID == 0 {
			data = append(data, item)
			continue
		}
		if matchGroup {
			data = append(data, item)
		}
	}
	var records []dto.ScriptInfo
	total, start, end := len(data), (req.Page-1)*req.PageSize, req.Page*req.PageSize
	if start > total {
		records = make([]dto.ScriptInfo, 0)
	} else {
		if end >= total {
			end = total
		}
		records = data[start:end]
	}
	return int64(total), records, nil
}

func (u *ScriptService) Create(req dto.ScriptOperate) error {
	itemData, _ := scriptRepo.Get(repo.WithByName(req.Name))
	if itemData.ID != 0 {
		return buserr.New("ErrRecordExist")
	}
	if err := copier.Copy(&itemData, &req); err != nil {
		return buserr.WithDetail("ErrStructTransform", err.Error(), nil)
	}
	if err := scriptRepo.Create(&itemData); err != nil {
		return err
	}
	if req.IsInteractive {
		return nil
	}
	if err := xpack.Sync(constant.SyncScripts); err != nil {
		global.LOG.Errorf("sync scripts to node failed, err: %v", err)
	}
	return nil
}

func (u *ScriptService) Delete(req dto.OperateByIDs) error {
	for _, item := range req.IDs {
		scriptItem, _ := scriptRepo.Get(repo.WithByID(item))
		if scriptItem.ID == 0 || scriptItem.IsSystem {
			continue
		}
		if err := scriptRepo.Delete(repo.WithByID(item)); err != nil {
			return err
		}
	}
	if err := xpack.Sync(constant.SyncScripts); err != nil {
		global.LOG.Errorf("sync scripts to node failed, err: %v", err)
	}
	return nil
}

func (u *ScriptService) Update(req dto.ScriptOperate) error {
	itemData, _ := scriptRepo.Get(repo.WithByID(req.ID))
	if itemData.ID == 0 {
		return buserr.New("ErrRecordNotFound")
	}
	updateMap := make(map[string]interface{})
	updateMap["name"] = req.Name
	updateMap["script"] = req.Script
	updateMap["groups"] = req.Groups
	updateMap["is_interactive"] = req.IsInteractive
	updateMap["description"] = req.Description
	if err := scriptRepo.Update(req.ID, updateMap); err != nil {
		return err
	}
	if err := xpack.Sync(constant.SyncScripts); err != nil {
		global.LOG.Errorf("sync scripts to node failed, err: %v", err)
	}
	return nil
}

func LoadScriptInfo(id uint) (model.ScriptLibrary, error) {
	return scriptRepo.Get(repo.WithByID(id))
}

func (u *ScriptService) Sync(req dto.OperateByTaskID) error {
	if global.CONF.Base.IsOffLine {
		return nil
	}
	syncTask, err := task.NewTaskWithOps(i18n.GetMsgByKey("ScriptLibrary"), task.TaskSync, task.TaskScopeScript, req.TaskID, 0)
	if err != nil {
		global.LOG.Errorf("create sync task failed %v", err)
		return err
	}

	syncTask.AddSubTask(task.GetTaskName(i18n.GetMsgByKey("ScriptLibrary"), task.TaskSync, task.TaskScopeScript), func(t *task.Task) (err error) {
		versionUrl := fmt.Sprintf("%s/scripts/version.txt", global.CONF.RemoteURL.ResourceURL)
		_, versionRes, err := req_helper.HandleRequestWithProxy(versionUrl, http.MethodGet, constant.TimeOut20s)
		if err != nil {
			return fmt.Errorf("load scripts version from remote failed, err: %v", err)
		}
		var scriptSetting model.Setting
		_ = global.DB.Where("key = ?", "ScriptVersion").First(&scriptSetting).Error
		localVersion := strings.ReplaceAll(string(versionRes), "\n", "")
		remoteVersion := strings.ReplaceAll(scriptSetting.Value, "\n", "")

		if localVersion == remoteVersion {
			syncTask.Log(i18n.GetMsgByKey("ScriptSyncSkip"))
			return nil
		}

		dataUrl := fmt.Sprintf("%s/scripts/data.yaml", global.CONF.RemoteURL.ResourceURL)
		_, dataRes, err := req_helper.HandleRequestWithProxy(dataUrl, http.MethodGet, constant.TimeOut20s)
		syncTask.LogWithStatus(i18n.GetMsgByKey("DownloadData"), err)
		if err != nil {
			return fmt.Errorf("load scripts data.yaml from remote failed, err: %v", err)
		}

		var scripts Scripts
		if err = yaml.Unmarshal(dataRes, &scripts); err != nil {
			return fmt.Errorf("the format of data.yaml is err: %v", err)
		}

		tmpDir := path.Join(global.CONF.Base.InstallDir, "1panel/tmp/script")
		if _, err := os.Stat(tmpDir); err != nil {
			_ = os.MkdirAll(tmpDir, 0755)
		}
		scriptsUrl := fmt.Sprintf("%s/scripts/scripts.tar.gz", global.CONF.RemoteURL.ResourceURL)
		err = files.DownloadFileWithProxy(scriptsUrl, tmpDir+"/scripts.tar.gz")
		syncTask.LogWithStatus(i18n.GetMsgByKey("DownloadPackage"), err)
		if err != nil {
			return fmt.Errorf("download scripts.tar.gz failed, err: %v", err)
		}

		if err := files.HandleUnTar(tmpDir+"/scripts.tar.gz", tmpDir, ""); err != nil {
			return fmt.Errorf("handle decompress scripts.tar.gz failed, err: %v", err)
		}
		var scriptsForDB []model.ScriptLibrary
		for _, item := range scripts.Scripts.Sh {
			itemName, _ := json.Marshal(item.Name)
			itemDescription, _ := json.Marshal(item.Description)
			shell, _ := os.ReadFile(fmt.Sprintf("%s/scripts/sh/%s.sh", tmpDir, item.Key))
			scriptItem := model.ScriptLibrary{
				Name:          string(itemName),
				IsInteractive: item.Interactive,
				IsSystem:      true,
				Script:        string(shell),
				Description:   string(itemDescription),
			}
			scriptsForDB = append(scriptsForDB, scriptItem)
		}

		syncTask.Log(i18n.GetMsgByKey("AnalyticCompletion"))
		if err := scriptRepo.SyncAll(scriptsForDB); err != nil {
			return fmt.Errorf("sync script with db failed, err: %v", err)
		}
		_ = os.RemoveAll(tmpDir)
		if err := global.DB.Model(&model.Setting{}).Where("key = ?", "ScriptVersion").Updates(map[string]interface{}{"value": string(versionRes)}).Error; err != nil {
			return fmt.Errorf("update script version in db failed, err: %v", err)
		}
		if err := xpack.Sync(constant.SyncScripts); err != nil {
			global.LOG.Errorf("sync scripts to node failed, err: %v", err)
		}
		return nil
	}, nil)

	if err := syncTask.Execute(); err != nil {
		return fmt.Errorf("sync scripts from remote failed, err: %v", err)
	}
	return nil
}

type Scripts struct {
	Scripts ScriptDetail `json:"scripts"`
}

type ScriptDetail struct {
	Sh []ScriptHelper `json:"sh"`
}

type ScriptHelper struct {
	Key         string            `json:"key"`
	Sort        uint              `json:"sort"`
	Groups      string            `json:"groups"`
	Name        map[string]string `json:"name"`
	Interactive bool              `json:"interactive"`
	Description map[string]string `json:"description"`
}
