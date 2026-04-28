package service

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

const clawhubGlobalRegistry = "https://clawhub.com"
const clawhubChinaRegistry = "https://mirror-cn.clawhub.com"

type openclawSkillsList struct {
	Skills []openclawSkillListItem `json:"skills"`
}

type openclawSkillListItem struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Source      string `json:"source"`
	Bundled     bool   `json:"bundled"`
	Disabled    bool   `json:"disabled"`
}

type openclawSkillInfo struct {
	SkillKey string `json:"skillKey"`
}

type skillhubSearchPayload struct {
	Skills  []dto.AgentSkillSearchItem `json:"skills"`
	Results []dto.AgentSkillSearchItem `json:"results"`
}

var clawhubSearchLinePattern = regexp.MustCompile(`^(\S+)\s+(.+?)\s+\(([\d.]+)\)$`)
var ansiEscapePattern = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)

func (a AgentService) ListSkills(req dto.AgentIDReq) ([]dto.AgentSkillItem, error) {
	agent, install, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return nil, err
	}
	if err := ensureContainerRunning(install.ContainerName); err != nil {
		return nil, err
	}
	if agent.AgentType == constant.AppHermesAgent {
		return listHermesSkills(install.ContainerName)
	}
	if agent.AgentType != constant.AppOpenclaw {
		return nil, fmt.Errorf("%s does not support", agent.AgentType)
	}
	output, err := cmd.RunDockerExecWithStdout(5*time.Minute, install.ContainerName, "openclaw", "skills", "list", "--json")
	if err != nil {
		return nil, err
	}
	if len(output) == 0 {
		return nil, nil
	}
	return parseOpenclawSkillsList(output)
}

func (a AgentService) SearchSkills(req dto.AgentSkillSearchReq) ([]dto.AgentSkillSearchItem, error) {
	agent, install, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return nil, err
	}
	if err := ensureContainerRunning(install.ContainerName); err != nil {
		return nil, err
	}
	if agent.AgentType == constant.AppHermesAgent {
		return searchHermesSkills(install.ContainerName, req.Source, req.Keyword)
	}
	if agent.AgentType != constant.AppOpenclaw {
		return nil, fmt.Errorf("%s does not support", agent.AgentType)
	}
	output, err := loadOpenclawSkillSearchOutput(install.ContainerName, req.Source, req.Keyword)
	if err != nil {
		return nil, err
	}
	if len(output) == 0 {
		return nil, nil
	}
	switch req.Source {
	case "skillhub":
		return parseSkillhubSearchResult(output)
	default:
		return parseClawhubSearchResult(output, req.Source), nil
	}
}

func (a AgentService) UpdateSkill(req dto.AgentSkillUpdateReq) error {
	agent, install, err := a.loadOpenclawAgentAndInstall(req.AgentID)
	if err != nil {
		return err
	}
	if err := ensureContainerRunning(install.ContainerName); err != nil {
		return err
	}
	conf, err := readOpenclawConfig(agent.ConfigPath)
	if err != nil {
		return err
	}
	skillKey, err := getOpenclawSkillKey(install.ContainerName, req.Name)
	if err != nil {
		return err
	}
	setOpenclawSkillEnabled(conf, skillKey, req.Enabled)
	return writeOpenclawConfigRaw(agent.ConfigPath, conf)
}

func (a AgentService) InstallSkill(req dto.AgentSkillInstallReq) error {
	agent, install, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return err
	}
	if err := ensureContainerRunning(install.ContainerName); err != nil {
		return err
	}
	installTask, err := task.NewTaskWithOps(req.Slug, task.TaskInstall, task.TaskScopeAI, req.TaskID, req.AgentID)
	if err != nil {
		return err
	}
	if agent.AgentType == constant.AppHermesAgent {
		installTask.AddSubTask("Install Hermes skill", func(t *task.Task) error {
			mgr := cmd.NewCommandMgr(cmd.WithTask(*t), cmd.WithContext(t.TaskCtx), cmd.WithTimeout(20*time.Minute))
			return mgr.Run("docker", buildHermesDockerExecArgs(install.ContainerName, "skills", "install", req.Slug, "--yes")...)
		}, nil)
		go func() {
			if err := installTask.Execute(); err != nil {
				global.LOG.Errorf("install hermes skill failed: %v", err)
			}
		}()
		return nil
	}
	if agent.AgentType != constant.AppOpenclaw {
		return fmt.Errorf("%s does not support", agent.AgentType)
	}
	installTask.AddSubTask("Install OpenClaw skill", func(t *task.Task) error {
		mgr := cmd.NewCommandMgr(cmd.WithTask(*t), cmd.WithContext(t.TaskCtx), cmd.WithTimeout(20*time.Minute))
		return installOpenclawSkill(mgr, install.ContainerName, req.Source, req.Slug)
	}, nil)
	go func() {
		if err := installTask.Execute(); err != nil {
			global.LOG.Errorf("install openclaw skill failed: %v", err)
		}
	}()
	return nil
}

func (a AgentService) UninstallSkill(req dto.AgentSkillUninstallReq) error {
	agent, install, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return err
	}
	if agent.AgentType != constant.AppHermesAgent {
		return fmt.Errorf("%s does not support", agent.AgentType)
	}
	if err := ensureContainerRunning(install.ContainerName); err != nil {
		return err
	}
	return cmd.NewCommandMgr(cmd.WithTimeout(5*time.Minute)).Run(
		"docker",
		buildHermesDockerExecCommandArgs(
			install.ContainerName,
			"sh",
			"-lc",
			fmt.Sprintf(`printf 'y\n' | %s skills uninstall "$1"`, hermesExecutablePath),
			"sh",
			req.Name,
		)...,
	)
}

func parseOpenclawSkillsList(output string) ([]dto.AgentSkillItem, error) {
	payloadBytes, err := extractEmbeddedJSON(output)
	if err != nil {
		return nil, err
	}
	if len(payloadBytes) == 0 {
		return nil, nil
	}
	var payload openclawSkillsList
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, err
	}
	items := make([]dto.AgentSkillItem, 0, len(payload.Skills))
	for _, item := range payload.Skills {
		items = append(items, dto.AgentSkillItem{
			Name:        item.Name,
			Description: item.Description,
			Source:      item.Source,
			Bundled:     item.Bundled,
			Disabled:    item.Disabled,
		})
	}
	return items, nil
}

func loadOpenclawSkillSearchOutput(containerName, source, keyword string) (string, error) {
	switch source {
	case "skillhub":
		return cmd.RunDockerExecWithStdout(2*time.Minute, containerName, "skillhub", "search", keyword, "--json")
	default:
		return cmd.RunDockerExecWithStdout(
			2*time.Minute,
			containerName,
			"sh",
			"-c",
			fmt.Sprintf("CLAWHUB_REGISTRY=%q clawhub search %q", resolveClawhubRegistry(source), keyword),
		)
	}
}

func parseSkillhubSearchResult(output string) ([]dto.AgentSkillSearchItem, error) {
	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		return nil, nil
	}
	var list []dto.AgentSkillSearchItem
	if err := json.Unmarshal([]byte(trimmed), &list); err == nil {
		for i := range list {
			list[i].Source = "skillhub"
		}
		return list, nil
	}
	var payload skillhubSearchPayload
	if err := json.Unmarshal([]byte(trimmed), &payload); err != nil {
		return nil, err
	}
	items := payload.Skills
	if len(items) == 0 {
		items = payload.Results
	}
	for i := range items {
		items[i].Source = "skillhub"
	}
	return items, nil
}

func parseClawhubSearchResult(output, source string) []dto.AgentSkillSearchItem {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	items := make([]dto.AgentSkillSearchItem, 0, len(lines))
	for _, line := range lines {
		matches := clawhubSearchLinePattern.FindStringSubmatch(strings.TrimSpace(line))
		if len(matches) != 4 {
			continue
		}
		items = append(items, dto.AgentSkillSearchItem{
			Slug:   matches[1],
			Name:   matches[2],
			Score:  matches[3],
			Source: source,
		})
	}
	return items
}

func installOpenclawSkill(mgr *cmd.CommandHelper, containerName, source, slug string) error {
	if err := mgr.Run("docker", "exec", containerName, "mkdir", "-p", openclawManagedSkillsDir); err != nil {
		return err
	}
	switch source {
	case "clawhub-global", "clawhub-cn":
		return mgr.Run(
			"docker",
			"exec",
			"-e",
			"CLAWHUB_REGISTRY="+resolveClawhubRegistry(source),
			containerName,
			"clawhub",
			"--workdir",
			"/home/node/.openclaw",
			"--dir",
			"skills",
			"install",
			slug,
		)
	default:
		return mgr.Run("docker", "exec", containerName, "skillhub", "--dir", openclawManagedSkillsDir, "install", slug)
	}
}

func resolveClawhubRegistry(source string) string {
	switch source {
	case "clawhub-cn":
		return clawhubChinaRegistry
	default:
		return clawhubGlobalRegistry
	}
}

func getOpenclawSkillKey(containerName, name string) (string, error) {
	output, err := cmd.RunDockerExecWithStdout(2*time.Minute, containerName, "openclaw", "skills", "info", name, "--json")
	if err != nil {
		return "", err
	}
	return parseOpenclawSkillKey(name, output)
}

func parseOpenclawSkillKey(name, output string) (string, error) {
	payloadBytes, err := extractEmbeddedJSON(output)
	if err != nil {
		return "", err
	}
	var payload openclawSkillInfo
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return "", err
	}
	if payload.SkillKey == "" {
		return "", fmt.Errorf("skill %s does not have a skillKey", name)
	}
	return payload.SkillKey, nil
}

func extractEmbeddedJSON(output string) ([]byte, error) {
	trimmed := strings.TrimSpace(ansiEscapePattern.ReplaceAllString(output, ""))
	if trimmed == "" {
		return nil, nil
	}
	for i := 0; i < len(trimmed); i++ {
		if trimmed[i] != '{' && trimmed[i] != '[' {
			continue
		}
		decoder := json.NewDecoder(strings.NewReader(trimmed[i:]))
		var raw json.RawMessage
		if err := decoder.Decode(&raw); err == nil {
			return raw, nil
		}
	}
	return nil, fmt.Errorf("json payload not found")
}

func setOpenclawSkillEnabled(conf map[string]interface{}, skillKey string, enabled bool) {
	skills := ensureChildMap(conf, "skills")
	entries := ensureChildMap(skills, "entries")
	entry := ensureChildMap(entries, skillKey)
	entry["enabled"] = enabled
}
