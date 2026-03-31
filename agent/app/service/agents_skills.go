package service

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

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
	_, install, err := a.loadOpenclawAgentAndInstall(req.AgentID)
	if err != nil {
		return nil, err
	}
	if err := ensureContainerRunning(install.ContainerName); err != nil {
		return nil, err
	}
	output, err := runDockerExecWithStdout(30*time.Second, install.ContainerName, "sh", "-c", "openclaw skills list --json 2>&1")
	if err != nil {
		return nil, err
	}
	if len(output) == 0 {
		return nil, nil
	}
	return parseOpenclawSkillsList(output)
}

func (a AgentService) SearchSkills(req dto.AgentSkillSearchReq) ([]dto.AgentSkillSearchItem, error) {
	_, install, err := a.loadOpenclawAgentAndInstall(req.AgentID)
	if err != nil {
		return nil, err
	}
	if err := ensureContainerRunning(install.ContainerName); err != nil {
		return nil, err
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
		return parseClawhubSearchResult(output), nil
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
	_, install, err := a.loadOpenclawAgentAndInstall(req.AgentID)
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
	installTask.AddSubTask("Install OpenClaw skill", func(t *task.Task) error {
		mgr := cmd.NewCommandMgr(cmd.WithTask(*t), cmd.WithContext(t.TaskCtx), cmd.WithTimeout(10*time.Minute))
		return mgr.Run("docker", "exec", install.ContainerName, "sh", "-c", buildOpenclawSkillInstallCommand(req.Source, req.Slug))
	}, nil)
	go func() {
		if err := installTask.Execute(); err != nil {
			global.LOG.Errorf("install openclaw skill failed: %v", err)
		}
	}()
	return nil
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
		return runDockerExecWithStdout(30*time.Second, containerName, "skillhub", "search", keyword, "--json")
	default:
		return runDockerExecWithStdout(30*time.Second, containerName, "clawhub", "search", keyword)
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

func parseClawhubSearchResult(output string) []dto.AgentSkillSearchItem {
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
			Source: "clawhub",
		})
	}
	return items
}

func buildOpenclawSkillInstallCommand(source, slug string) string {
	switch source {
	case "clawhub":
		return fmt.Sprintf(
			"mkdir -p %s && clawhub --workdir /home/node/.openclaw --dir skills install %q",
			openclawManagedSkillsDir,
			slug,
		)
	default:
		return fmt.Sprintf(
			"mkdir -p %s && skillhub --dir %s install %q",
			openclawManagedSkillsDir,
			openclawManagedSkillsDir,
			slug,
		)
	}
}

func getOpenclawSkillKey(containerName, name string) (string, error) {
	output, err := runDockerExecWithStdout(30*time.Second, containerName, "sh", "-c", fmt.Sprintf("openclaw skills info %q --json 2>&1", name))
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
