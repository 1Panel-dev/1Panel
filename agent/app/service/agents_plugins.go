package service

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/compose"
)

var (
	openclawPluginPackagePattern = regexp.MustCompile(`^(@[a-z0-9][a-z0-9._-]*/)?[a-z0-9][a-z0-9._-]*$`)
	openclawPluginVersionPattern = regexp.MustCompile(`^[0-9A-Za-z][0-9A-Za-z._-]*$`)
	openclawPluginIDPattern      = regexp.MustCompile(`^(@[A-Za-z0-9][A-Za-z0-9._-]*/)?[A-Za-z0-9][A-Za-z0-9._-]*$`)
)

type openclawPluginListOutput struct {
	Plugins []struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Version string `json:"version"`
		Origin  string `json:"origin"`
		Enabled bool   `json:"enabled"`
	} `json:"plugins"`
}

type openclawPluginIndexItem struct {
	PluginID       string `json:"pluginId"`
	PackageName    string `json:"packageName"`
	PackageVersion string `json:"packageVersion"`
	Origin         string `json:"origin"`
	Enabled        bool   `json:"enabled"`
}

type openclawPluginSearchOutput struct {
	Results []struct {
		Score   float64 `json:"score"`
		Package struct {
			Name             string   `json:"name"`
			RuntimeID        string   `json:"runtimeId"`
			DisplayName      string   `json:"displayName"`
			Summary          string   `json:"summary"`
			LatestVersion    string   `json:"latestVersion"`
			Categories       []string `json:"categories"`
			Channel          string   `json:"channel"`
			IsOfficial       bool     `json:"isOfficial"`
			VerificationTier string   `json:"verificationTier"`
			Stats            struct {
				Downloads int64 `json:"downloads"`
			} `json:"stats"`
		} `json:"package"`
	} `json:"results"`
}

func (a AgentService) ListPlugins(req dto.AgentPluginsReq) ([]dto.AgentPluginItem, error) {
	agent, install, err := a.loadOpenclawAgentAndInstall(req.AgentID)
	if err != nil {
		return nil, err
	}
	if plugins, err := readOpenclawPluginIndex(filepath.Join(filepath.Dir(agent.ConfigPath), "state", "openclaw.sqlite")); err == nil {
		return plugins, nil
	}
	output, err := cmd.RunDockerExecWithStdout(2*time.Minute, install.ContainerName, "openclaw", "plugins", "list", "--json")
	if err != nil {
		return nil, err
	}
	return parseOpenclawPluginList([]byte(output))
}

func (a AgentService) SearchPlugins(req dto.AgentPluginSearchReq) ([]dto.AgentPluginSearchItem, error) {
	_, install, err := a.loadOpenclawAgentAndInstall(req.AgentID)
	if err != nil {
		return nil, err
	}
	limit := req.Limit
	if limit == 0 {
		limit = 20
	}
	output, err := cmd.RunDockerExecWithStdout(
		2*time.Minute,
		install.ContainerName,
		"openclaw", "plugins", "search", strings.TrimSpace(req.Keyword), "--limit", fmt.Sprint(limit), "--json",
	)
	if err != nil {
		return nil, err
	}
	return parseOpenclawPluginSearch([]byte(output))
}

func (a AgentService) InstallMarketPlugin(req dto.AgentPluginMarketInstallReq) error {
	spec, err := buildOpenclawPluginInstallSpec(req.Package, req.Version)
	if err != nil {
		return err
	}
	_, install, err := a.loadOpenclawAgentAndInstall(req.AgentID)
	if err != nil {
		return err
	}
	if err := task.CheckScopeTaskIsExecuting(task.TaskScopeAI, req.AgentID); err != nil {
		return err
	}
	taskName := fmt.Sprintf("%s [%s]", i18n.GetMsgByKey("AgentPluginInstall"), req.Package)
	installTask, err := task.NewTask(taskName, task.TaskInstall, task.TaskScopeAI, req.TaskID, req.AgentID)
	if err != nil {
		return err
	}
	installTask.AddSubTask(taskName, func(t *task.Task) error {
		mgr := cmd.NewCommandMgr(cmd.WithTask(*t), cmd.WithContext(t.TaskCtx), cmd.WithTimeout(10*time.Minute))
		return mgr.Run("docker", "exec", install.ContainerName, "openclaw", "plugins", "install", spec)
	}, nil)
	addOpenclawPluginRestartTask(installTask, install)
	go executeAgentPluginTask(installTask)
	return nil
}

func (a AgentService) OperatePlugin(req dto.AgentPluginOperateReq) error {
	if !openclawPluginIDPattern.MatchString(req.PluginID) {
		return buserr.New("ErrInvalidChar")
	}
	agent, install, err := a.loadOpenclawAgentAndInstall(req.AgentID)
	if err != nil {
		return err
	}
	if err := task.CheckScopeTaskIsExecuting(task.TaskScopeAI, req.AgentID); err != nil {
		return err
	}
	if req.Operate == "update" || req.Operate == "uninstall" {
		plugins, err := a.ListPlugins(dto.AgentPluginsReq{AgentID: req.AgentID})
		if err != nil {
			return err
		}
		for _, plugin := range plugins {
			if plugin.ID == req.PluginID && plugin.Origin == "bundled" {
				return buserr.WithName("ErrNotSupportType", req.Operate)
			}
		}
	}

	taskType := map[string]string{
		"enable":    task.TaskUpdate,
		"disable":   task.TaskUpdate,
		"update":    task.TaskUpgrade,
		"uninstall": task.TaskUninstall,
	}[req.Operate]
	taskName := fmt.Sprintf("%s [%s]", i18n.GetMsgByKey(map[string]string{
		"enable":    "AgentPluginEnable",
		"disable":   "AgentPluginDisable",
		"update":    "AgentPluginUpdate",
		"uninstall": "AgentPluginUninstall",
	}[req.Operate]), req.PluginID)
	operateTask, err := task.NewTask(taskName, taskType, task.TaskScopeAI, req.TaskID, req.AgentID)
	if err != nil {
		return err
	}
	operateTask.AddSubTask(taskName, func(t *task.Task) error {
		mgr := cmd.NewCommandMgr(cmd.WithTask(*t), cmd.WithContext(t.TaskCtx), cmd.WithTimeout(10*time.Minute))
		if req.Operate == "uninstall" {
			if err := uninstallOpenclawPlugin(mgr, install.ContainerName, req.PluginID); err != nil {
				return err
			}
			return cleanupManagedOpenclawPlugin(agent, req.PluginID)
		}
		return mgr.Run("docker", "exec", install.ContainerName, "openclaw", "plugins", req.Operate, req.PluginID)
	}, nil)
	addOpenclawPluginRestartTask(operateTask, install)
	go executeAgentPluginTask(operateTask)
	return nil
}

func parseOpenclawPluginList(raw []byte) ([]dto.AgentPluginItem, error) {
	payload, err := extractEmbeddedJSON(string(raw))
	if err != nil {
		return nil, err
	}
	if len(payload) == 0 {
		return []dto.AgentPluginItem{}, nil
	}
	var output openclawPluginListOutput
	if err := json.Unmarshal(payload, &output); err != nil {
		return nil, err
	}
	items := make([]dto.AgentPluginItem, 0, len(output.Plugins))
	for _, plugin := range output.Plugins {
		items = append(items, dto.AgentPluginItem{
			ID:      plugin.ID,
			Name:    plugin.Name,
			Version: plugin.Version,
			Origin:  plugin.Origin,
			Enabled: plugin.Enabled,
		})
	}
	return items, nil
}

func readOpenclawPluginIndex(dbPath string) ([]dto.AgentPluginItem, error) {
	db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(dbPath)+"?mode=ro")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var raw []byte
	if err := db.QueryRow(
		"SELECT plugins_json FROM installed_plugin_index WHERE index_key = ?",
		"installed-plugin-index",
	).Scan(&raw); err != nil {
		return nil, err
	}

	var plugins []openclawPluginIndexItem
	if err := json.Unmarshal(raw, &plugins); err != nil {
		return nil, err
	}
	items := make([]dto.AgentPluginItem, 0, len(plugins))
	for _, plugin := range plugins {
		name := plugin.PackageName
		if name == "" {
			name = plugin.PluginID
		}
		items = append(items, dto.AgentPluginItem{
			ID:      plugin.PluginID,
			Name:    name,
			Version: plugin.PackageVersion,
			Origin:  plugin.Origin,
			Enabled: plugin.Enabled,
		})
	}
	return items, nil
}

func parseOpenclawPluginSearch(raw []byte) ([]dto.AgentPluginSearchItem, error) {
	payload, err := extractEmbeddedJSON(string(raw))
	if err != nil {
		return nil, err
	}
	if len(payload) == 0 {
		return []dto.AgentPluginSearchItem{}, nil
	}
	var output openclawPluginSearchOutput
	if err := json.Unmarshal(payload, &output); err != nil {
		return nil, err
	}
	items := make([]dto.AgentPluginSearchItem, 0, len(output.Results))
	for _, result := range output.Results {
		items = append(items, dto.AgentPluginSearchItem{
			Package:          result.Package.Name,
			PluginID:         result.Package.RuntimeID,
			Name:             result.Package.DisplayName,
			Description:      result.Package.Summary,
			Version:          result.Package.LatestVersion,
			Channel:          result.Package.Channel,
			VerificationTier: result.Package.VerificationTier,
			Categories:       append([]string{}, result.Package.Categories...),
			Official:         result.Package.IsOfficial,
			Downloads:        result.Package.Stats.Downloads,
			Score:            result.Score,
		})
	}
	return items, nil
}

func buildOpenclawPluginInstallSpec(packageName, version string) (string, error) {
	packageName = strings.TrimSpace(packageName)
	version = strings.TrimSpace(version)
	if !openclawPluginPackagePattern.MatchString(packageName) || !openclawPluginVersionPattern.MatchString(version) {
		return "", buserr.New("ErrInvalidChar")
	}
	return "clawhub:" + packageName + "@" + version, nil
}

func cleanupManagedOpenclawPlugin(agent *model.Agent, pluginID string) error {
	pluginType := map[string]string{
		"openclaw-lark":         "feishu",
		"openclaw-qqbot":        "qqbot",
		"wecom-openclaw-plugin": "wecom",
		"dingtalk-connector":    "dingtalk",
		"openclaw-weixin":       "weixin",
	}[pluginID]
	if pluginType == "" {
		return nil
	}
	conf, err := readOpenclawConfig(agent.ConfigPath)
	if err != nil {
		return err
	}
	cleanupOpenclawPluginConfig(conf, pluginType)
	return writeOpenclawConfigRaw(agent.ConfigPath, conf)
}

func addOpenclawPluginRestartTask(t *task.Task, install *model.AppInstall) {
	t.AddSubTask(task.GetTaskName("OpenClaw", task.TaskRestart, task.TaskScopeAI), func(t *task.Task) error {
		output, err := compose.Restart(install.GetComposePath())
		if output != "" {
			t.Log(output)
		}
		return err
	}, nil)
}

func executeAgentPluginTask(t *task.Task) {
	if err := t.Execute(); err != nil {
		global.LOG.Errorf("operate openclaw plugin failed: %v", err)
	}
}
