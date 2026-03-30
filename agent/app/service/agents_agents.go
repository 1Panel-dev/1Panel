package service

import (
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

var agentMarkdownFileNames = []string{"AGENTS.md", "SOUL.md", "USER.md", "IDENTITY.md", "TOOLS.md", "HEARTBEAT.md", "BOOT.md", "BOOTSTRAP.md"}

func (a AgentService) CreateRole(req dto.AgentRoleCreateReq) (*dto.AgentRoleCreateResp, error) {
	_, install, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return nil, err
	}

	name := strings.TrimSpace(req.Name)
	args := []string{"exec", install.ContainerName, "openclaw", "agents", "add", name}
	workspace := "/home/node/.openclaw/workspace-agent_" + name
	agentDir := "/home/node/.openclaw/agents/" + name
	args = append(args, "--workspace", workspace)
	if model := strings.TrimSpace(req.Model); model != "" {
		args = append(args, "--model", model)
	}
	for _, binding := range req.Bindings {
		channel := strings.TrimSpace(binding.Channel)
		if channel == "" {
			continue
		}
		if accountID := strings.TrimSpace(binding.AccountID); accountID != "" {
			channel = channel + ":" + accountID
		}
		args = append(args, "--bind", channel)
	}
	args = append(args, "--agent-dir", agentDir)
	args = append(args, "--non-interactive", "--json")

	mgr := cmd.NewCommandMgr(cmd.WithTimeout(2 * time.Minute))
	output, err := mgr.RunWithStdout("docker", args...)
	if err != nil {
		return nil, err
	}
	return &dto.AgentRoleCreateResp{Output: strings.TrimSpace(output)}, nil
}

func (a AgentService) GetConfiguredAgents(req dto.AgentConfiguredAgentsReq) ([]dto.AgentConfiguredAgentItem, error) {
	agent, _, conf, err := a.loadAgentConfig(req.AgentID)
	if err != nil {
		return nil, err
	}
	agents, ok := conf["agents"].(map[string]interface{})
	if !ok {
		return []dto.AgentConfiguredAgentItem{}, nil
	}
	rawList, ok := agents["list"].([]interface{})
	if !ok {
		return []dto.AgentConfiguredAgentItem{}, nil
	}
	result := make([]dto.AgentConfiguredAgentItem, 0, len(rawList))
	baseDir := path.Join(global.Dir.AppInstallDir, agent.AgentType, agent.Name, "data")
	for _, item := range rawList {
		record, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		configured := extractConfiguredAgentItem(baseDir, record)
		if strings.EqualFold(configured.ID, "main") {
			continue
		}
		result = append(result, configured)
	}
	applyConfiguredAgentBindings(result, conf["bindings"])
	return result, nil
}

func (a AgentService) GetRoleChannels(req dto.AgentRoleChannelsReq) ([]dto.AgentRoleChannelItem, error) {
	_, _, conf, err := a.loadAgentConfig(req.AgentID)
	if err != nil {
		return nil, err
	}
	channels, ok := conf["channels"].(map[string]interface{})
	if !ok || len(channels) == 0 {
		return []dto.AgentRoleChannelItem{}, nil
	}
	boundChannels := loadBoundChannelSet(conf["bindings"])
	result := make([]dto.AgentRoleChannelItem, 0, len(channels))
	for key := range channels {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		channelConf, _ := channels[key].(map[string]interface{})
		result = append(result, dto.AgentRoleChannelItem{
			Name:       key,
			Bound:      boundChannels[key],
			AccountIDs: extractChannelAccountIDs(channelConf),
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result, nil
}

func (a AgentService) DeleteRole(req dto.AgentRoleDeleteReq) error {
	agent, install, conf, err := a.loadAgentConfig(req.AgentID)
	if err != nil {
		return err
	}

	baseDir := path.Join(global.Dir.AppInstallDir, agent.AgentType, agent.Name, "data")
	roleID := strings.TrimSpace(req.ID)
	if roleID == "" {
		return buserr.New("ErrRecordNotFound")
	}
	target, ok := findConfiguredAgentByID(baseDir, conf, roleID)
	if !ok {
		return buserr.New("ErrRecordNotFound")
	}

	args := []string{"exec", install.ContainerName, "openclaw", "agents", "delete", roleID, "--force"}

	mgr := cmd.NewCommandMgr(cmd.WithTimeout(2 * time.Minute))
	if _, err = mgr.RunWithStdout("docker", args...); err != nil {
		return err
	}
	if target.Workspace != "" {
		if err := os.RemoveAll(target.Workspace); err != nil {
			return err
		}
	}
	if target.AgentDir != "" {
		if err := os.RemoveAll(target.AgentDir); err != nil {
			return err
		}
	}
	return nil
}

func (a AgentService) GetRoleMarkdownFiles(req dto.AgentRoleMarkdownFilesReq) ([]dto.AgentRoleMarkdownFileItem, error) {
	agent, err := loadOpenclawAgentByID(req.AgentID)
	if err != nil {
		return nil, err
	}

	baseDir := path.Join(global.Dir.AppInstallDir, agent.AgentType, agent.Name, "data")
	workspaceDir, err := resolveMarkdownWorkspaceDir(baseDir, req.Workspace)
	if err != nil {
		return nil, err
	}

	items := make([]dto.AgentRoleMarkdownFileItem, 0, len(agentMarkdownFileNames))
	for _, name := range agentMarkdownFileNames {
		item := dto.AgentRoleMarkdownFileItem{
			Name: name,
		}
		content, readErr := os.ReadFile(path.Join(workspaceDir, name))
		if readErr == nil {
			item.Content = string(content)
		} else if !os.IsNotExist(readErr) {
			return nil, readErr
		}
		items = append(items, item)
	}
	return items, nil
}

func (a AgentService) UpdateRoleMarkdownFiles(req dto.AgentRoleMarkdownFilesUpdateReq) error {
	agent, install, err := a.loadOpenclawAgentAndInstall(req.AgentID)
	if err != nil {
		return err
	}

	baseDir := path.Join(global.Dir.AppInstallDir, agent.AgentType, agent.Name, "data")
	dirPath, err := resolveMarkdownWorkspaceDir(baseDir, req.Workspace)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(dirPath, 0755); err != nil {
		return err
	}
	for _, item := range req.Files {
		file := path.Join(dirPath, item.Name)
		if err = os.WriteFile(file, []byte(item.Content), 0644); err != nil {
			return err
		}
	}
	if req.Restart {
		return NewIAppInstalledService().Operate(request.AppInstalledOperate{
			InstallId: install.ID,
			Operate:   constant.Restart,
		})
	}
	return nil
}

func extractConfiguredAgentItem(installDir string, record map[string]interface{}) dto.AgentConfiguredAgentItem {
	item := dto.AgentConfiguredAgentItem{Bindings: []dto.AgentRoleBinding{}}
	if id, ok := record["id"].(string); ok {
		item.ID = strings.TrimSpace(id)
	} else if id, ok := record["agentId"].(string); ok {
		item.ID = strings.TrimSpace(id)
	}
	if name, ok := record["name"].(string); ok {
		item.Name = strings.TrimSpace(name)
	}
	if workspace, ok := record["workspace"].(string); ok {
		item.Workspace = resolveRoleDir(installDir, strings.TrimSpace(workspace))
	}
	if model, ok := record["model"].(string); ok {
		item.Model = strings.TrimSpace(model)
	}
	if agentDir, ok := record["agentDir"].(string); ok {
		item.AgentDir = strings.TrimSpace(agentDir)
	} else if agentDir, ok := record["agent_dir"].(string); ok {
		item.AgentDir = strings.TrimSpace(agentDir)
	}
	item.AgentDir = resolveRoleDir(installDir, item.AgentDir)
	return item
}

func findConfiguredAgentByID(baseDir string, conf map[string]interface{}, id string) (dto.AgentConfiguredAgentItem, bool) {
	agents, ok := conf["agents"].(map[string]interface{})
	if !ok {
		return dto.AgentConfiguredAgentItem{}, false
	}
	rawList, ok := agents["list"].([]interface{})
	if !ok {
		return dto.AgentConfiguredAgentItem{}, false
	}
	for _, item := range rawList {
		record, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		configured := extractConfiguredAgentItem(baseDir, record)
		if strings.EqualFold(strings.TrimSpace(configured.ID), strings.TrimSpace(id)) {
			return configured, true
		}
	}
	return dto.AgentConfiguredAgentItem{}, false
}

func applyConfiguredAgentBindings(agents []dto.AgentConfiguredAgentItem, value interface{}) {
	bindings, ok := value.([]interface{})
	if !ok || len(agents) == 0 {
		return
	}
	indexByID := make(map[string]int, len(agents))
	indexByName := make(map[string]int, len(agents))
	for i, agent := range agents {
		if agent.ID != "" {
			indexByID[agent.ID] = i
		}
		if agent.Name != "" {
			indexByName[agent.Name] = i
		}
	}
	for _, binding := range bindings {
		record, ok := binding.(map[string]interface{})
		if !ok {
			continue
		}
		if bindingType, _ := record["type"].(string); !strings.EqualFold(strings.TrimSpace(bindingType), "route") {
			continue
		}
		targetID, _ := record["agentId"].(string)
		targetID = strings.TrimSpace(targetID)
		if targetID == "" {
			continue
		}
		match, ok := record["match"].(map[string]interface{})
		if !ok {
			continue
		}
		channel, _ := match["channel"].(string)
		channel = strings.TrimSpace(channel)
		if channel == "" {
			continue
		}
		index, ok := indexByID[targetID]
		if !ok {
			index, ok = indexByName[targetID]
		}
		if !ok {
			continue
		}
		accountID, _ := match["accountId"].(string)
		if strings.TrimSpace(accountID) == "" {
			accountID, _ = record["accountId"].(string)
		}
		agents[index].Bindings = append(agents[index].Bindings, dto.AgentRoleBinding{
			Channel:   channel,
			AccountID: strings.TrimSpace(accountID),
		})
	}
}

func loadBoundChannelSet(value interface{}) map[string]bool {
	result := make(map[string]bool)
	bindings, ok := value.([]interface{})
	if !ok {
		return result
	}
	for _, binding := range bindings {
		record, ok := binding.(map[string]interface{})
		if !ok {
			continue
		}
		if bindingType, _ := record["type"].(string); !strings.EqualFold(strings.TrimSpace(bindingType), "route") {
			continue
		}
		match, ok := record["match"].(map[string]interface{})
		if !ok {
			continue
		}
		channel, _ := match["channel"].(string)
		channel = strings.TrimSpace(channel)
		if channel == "" {
			continue
		}
		result[channel] = true
	}
	return result
}

func extractChannelAccountIDs(channel map[string]interface{}) []string {
	if len(channel) == 0 {
		return []string{}
	}
	accounts, ok := channel["accounts"].(map[string]interface{})
	if !ok || len(accounts) == 0 {
		return []string{}
	}
	result := make([]string, 0, len(accounts))
	for key := range accounts {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}

func resolveRoleDir(installDir, workspace string) string {
	workspace = strings.TrimSpace(workspace)
	if workspace == "" {
		return ""
	}
	if strings.HasPrefix(workspace, "/home/node/.openclaw/workspace/") {
		return strings.ReplaceAll(workspace, "/home/node/.openclaw", installDir)
	}
	return strings.ReplaceAll(workspace, "/home/node/.openclaw", path.Join(installDir, "conf"))
}

func resolveMarkdownWorkspaceDir(installDir, workspace string) (string, error) {
	workspace = strings.TrimSpace(workspace)
	if workspace == "" {
		return "", buserr.New("ErrRecordNotFound")
	}

	confDir := path.Join(installDir, "conf")
	allowedOpenclawPrefixes := []string{
		"/home/node/.openclaw/workspace/",
		"/home/node/.openclaw/workspace-agent_",
	}
	for _, prefix := range allowedOpenclawPrefixes {
		if strings.HasPrefix(workspace, prefix) {
			return resolveRoleDir(installDir, workspace), nil
		}
	}

	cleanWorkspace := path.Clean(workspace)
	cleanConfDir := path.Clean(confDir)
	if cleanWorkspace == cleanConfDir || strings.HasPrefix(cleanWorkspace, cleanConfDir+"/") {
		return cleanWorkspace, nil
	}
	return "", buserr.New("ErrRecordNotFound")
}
